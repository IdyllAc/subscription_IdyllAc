package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"

	_ "modernc.org/sqlite"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

var db *sql.DB

func main() {
 	// Load .env environment variables
 	godotenv.Load()

// Set SESSION_SECRET for Goth
key := os.Getenv("SESSION_SECRET")
maxAge := 86400 * 30 // 30 days
isProd := false       // change to true in production

store := sessions.NewCookieStore([]byte(key))
store.MaxAge(maxAge)
store.Options.Path = "/"
store.Options.HttpOnly = true
store.Options.Secure = isProd
gothic.Store = store


	// Serve static files (CSS, JS, Images, etc)
 fs := http.FileServer(http.Dir("./static"))
 http.Handle("/static/", http.StripPrefix("/static/", fs))


	// Initialize database
	var err error
	db, err = sql.Open("sqlite", "./db_subscribers.db") // ✅ Adjusted for modernc.org/sqlite
	if err != nil {
		log.Fatal("DB open error", err)
	}
	defer db.Close()

	// Create table if it doesn't exists
	createTable()


	// Setup OAuth Providers
	goth.UseProviders(
		facebook.New(
			"1414625536386549", // 👉 Replace with your Facebook App ID
			"34b3780d34e63750b0a2af27f52490e1", // 👉 Replace with your Facebook App Secret
			"http://localhost:8080/auth/facebook/callback",
		),
		google.New(
			"94664221445-u9fkhtf34koasqnrf93vakqfdoe4nitl.apps.googleusercontent.com", // 👉 Replace with your Google Client ID
			"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", // 👉 Replace with your Google Client Secret
			"http://localhost:8080/auth/google/callback",
			"email", "profile",
		),
		github.New(
			"Ov23lizsJT6GZblzvDPW",        // 👉 Replace with your real GitHub Client ID
			"168d33d988717c92ff783881837cc13a50095ec4",    // 👉 Replace with your real GitHub Client Secret
			"http://localhost:8080/auth/github/callback",
		),
	)

	// Setup Routes
  http.HandleFunc("/", serveIndex)
	http.HandleFunc("/subscribe", serveSubscribe)
	http.HandleFunc("/subscribe/email", handleEmailSubscription)

	http.HandleFunc("/subscribers", handleListSubscribers)

	http.HandleFunc("/view-emails", handleViewEmails)

	http.HandleFunc("/submit", handleFormSubmission)

	http.HandleFunc("/auth/facebook", handleFacebookLogin)
	http.HandleFunc("/auth/facebook/callback", handleFacebookCallback)

	http.HandleFunc("/auth/google", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

	http.HandleFunc("/auth/github", handleGitHubLogin)
 http.HandleFunc("/auth/github/callback", handleGitHubCallback)

	


	// Start server
	fmt.Println("✅ Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ✅ This function is now outside of main
func createTable() {
query := `
		CREATE TABLE IF NOT EXISTS subscribers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL
		);
		`
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("❌ Failed to create DB table: %v", err)
		}
	}


func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func serveSubscribe(w http.ResponseWriter, r *http.Request) {
	 // Ensure the method is GET to serve the subscription page
	if r.Method == http.MethodGet {
		 // Serve the subscribe.html file
		http.ServeFile(w, r, "subscribe.html")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func handleEmailSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")  // Get the email from the form submission
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	
	// Insert into database
	_, err := db.Exec("INSERT OR IGNORE INTO subscribers (email) VALUES (?)", email)
	if err != nil {
		log.Println("❌ DB error:", err)
		http.Error(w, "Failed to save email to database"+err.Error(), http.StatusInternalServerError)
		return
	}

// Save to .txt file
file, err := os.OpenFile("subscribers_emails.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("❌ File open error:", err)
		http.Error(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(email + "\n"); err != nil {
		log.Println("❌ File write error:", err)
		http.Error(w, "Failed to write to file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate verification link
	link := "http://localhost:8080/verify?email=" + url.QueryEscape(email)

	// Send confirmation email
	sendConfirmationEmail(email, link)

	// Final response to client
	fmt.Fprintf(w, "✅ Subscription successful! A confirmation has been sent to: %s", email)
	}

func sendConfirmationEmail(to string, link string) {
	from := "victor.via7@gmail.com"          // ✅ Your Gmail address
	password := "ewbr xtgv nlxi dxmy"        // ✅ App password (not Gmail login password)
	subject := "Verify your subscription"
	body := fmt.Sprintf("Click the link to verify your subscription:\n\n%s", link)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" + body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, password, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
	} else {
		log.Printf("✅ Confirmation email sent to %s", to)
	}
}

func handleViewEmails(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("./subscribe/emails.txt")
	if err != nil {
		http.Error(w, "Failed to read emails", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}


func handleListSubscribers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT email FROM subscribers")
	if err != nil {
		http.Error(w, "Failed to fetch subscribers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			continue
		}
		result += email + "\n"
	}

	w.Write([]byte("✅ Subscribers List:\n" + result))
}


func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=facebook"
	gothic.BeginAuthHandler(w, r)
}

func handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "Facebook Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "✅ Facebook Login Successful!\nName: %s\nEmail: %s\n", user.Name, user.Email)
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=google"
	gothic.BeginAuthHandler(w, r)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "Google Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "✅ Google Login Successful!\n\nName: %s\nEmail: %s\n", user.Name, user.Email)
}

func handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=github"
	gothic.BeginAuthHandler(w, r)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
					http.Error(w, "GitHub Login failed: "+err.Error(), http.StatusInternalServerError)
					return
	}
	fmt.Fprintf(w, "✅ GitHub Login Successful!\nName: %s\nEmail: %s\n", user.Name, user.Email)
}


func handleFormSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ParseForm() error: "+err.Error(), http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	message := r.FormValue("message")

	if email == "" || message == "" {
		http.Error(w, "Email and message are required", http.StatusBadRequest)
		return
	}

	// Example: Print to console
	fmt.Printf("✅ New message from %s: %s\n", email, message)

	// (Optional) Save to database or send as email...

	w.Write([]byte("✅ Thank you for your message!"))
}





