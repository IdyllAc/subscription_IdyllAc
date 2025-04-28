package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"

	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB


func main() {
	// Load environment variables
	godotenv.Load()

	// Serve static files (CSS, JS, Images, etc)
fs := http.FileServer(http.Dir("./static"))
http.Handle("/static/", http.StripPrefix("/static/", fs))


	// Initialize database
	var err error
	db, err = sql.Open("sqlite3", "./db_subscribers")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create subscribers table if not exists
	createTable()

	// Setup OAuth Providers
	goth.UseProviders(
		facebook.New(
			"842237368120640", // 👉 Replace with your Facebook App ID
			"02d25f9c8470d6835d10858bfa12b4c7", // 👉 Replace with your Facebook App Secret
			"http://localhost:8080/auth/facebook/callback",
		),
		google.New(
			"94664221445-u9fkhtf34koasqnrf93vakqfdoe4nitl.apps.googleusercontent.com", // 👉 Replace with your Google Client ID
			"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", // 👉 Replace with your Google Client Secret
			"http://localhost:8080/auth/google/callback",
			"email", "profile",
		),
	)

	// Routes
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/subscribe", serveSubscribe)
	http.HandleFunc("/subscribe/email", handleEmailSubscription)
	http.HandleFunc("/auth/facebook", handleFacebookLogin)
	http.HandleFunc("/auth/facebook/callback", handleFacebookCallback)
	http.HandleFunc("/auth/google", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

	// Start server
	fmt.Println("✅ Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "subscribe.html")
}

func handleEmailSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT OR IGNORE INTO subscribers (email) VALUES (?)", email)
	if err != nil {
		http.Error(w, "Failed to save email", http.StatusInternalServerError)
		log.Println("DB error:", err)
		return
	}

	// Generate verification link
	link := "http://localhost:8080/verify?email=" + url.QueryEscape(email)

	// Send confirmation email
	sendConfirmationEmail(email, link)

	fmt.Fprintf(w, "✅ Thanks! Subscription successful, confirmation sent to: %s", email)
}

func sendConfirmationEmail(to string, link string) {
	from := "victor.via7@gmail.com"          // ✅ Your Gmail address
	password := "feyu ndvj skxu wbxj"         // ✅ App password (not Gmail login password)
	subject := "Verify your subscription"
	body := fmt.Sprintf("Click the link to verify your subscription:\n\n%s", link)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" + body

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", from, password, "smtp.gmail.com"),
		from,
		[]string{to},
		[]byte(msg),
	)

	if err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
	} else {
		log.Printf("✅ Confirmation email sent to %s", to)
	}
}

func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	r.URL.RawQuery = "provider=facebook"
	gothic.BeginAuthHandler(w, r)
}

func handleFacebookCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, "Login failed: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "✅ Google Login Successful!\n\nName: %s\nEmail: %s\n", user.Name, user.Email)
}
