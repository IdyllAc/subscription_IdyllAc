// package main

// import (
//     "database/sql"
//     "context"
//     "encoding/json"
//     "fmt"
//     "log"
//     "net/http"
//      "os"
//     "net/smtp"
//     _ "github.com/mattn/go-sqlite3"
//     "golang.org/x/oauth2"
//     "golang.org/x/oauth2/google"
//     "google.golang.org/api/gmail/v1"

//      "github.com/markbates/goth"
//      "github.com/markbates/goth/gothic"
//      "github.com/markbates/goth/providers/facebook"
//      "github.com/markbates/goth/providers/google"

// )



// var (
//     oauthConfig *oauth2.Config
//     state       = "randomstring123" // احرص على تغييره في الإنتاج
// )

// func main() {
//     b, err := os.ReadFile("credentials.json")
//     if err != nil {
//         log.Fatalf("Unable to read client secret file: %v", err)
//     }

//     oauthConfig, err = google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
//     if err != nil {
//         log.Fatalf("Unable to parse client secret file to config: %v", err)
//     }

//     http.HandleFunc("/", handleMain)
//     http.HandleFunc("/auth", handleGmailLogin)
//     http.HandleFunc("/callback", handleGmailCallback)

//     fmt.Println("Server started at http://localhost:8080")
//     log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func handleMain(w http.ResponseWriter, r *http.Request) {
//     http.ServeFile(w, r, "index.html")
// }

// func handleGmailLogin(w http.ResponseWriter, r *http.Request) {
//     url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
//     http.Redirect(w, r, url, http.StatusTemporaryRedirect)
// }

// func handleGmailCallback(w http.ResponseWriter, r *http.Request) {
//     if r.FormValue("state") != state {
//         http.Error(w, "State mismatch", http.StatusBadRequest)
//         return
//     }

//     token, err := oauthConfig.Exchange(context.Background(), r.FormValue("code"))
//     if err != nil {
//         http.Error(w, "Code exchange failed", http.StatusInternalServerError)
//         return
//     }

//     client := oauthConfig.Client(context.Background(), token)
//     srv, err := gmail.New(client)
//     if err != nil {
//         log.Fatalf("Unable to retrieve Gmail client: %v", err)
//     }

//     user := "me"
//     rList, err := srv.Users.Messages.List(user).Do()
//     if err != nil {
//         log.Fatalf("Unable to retrieve messages: %v", err)
//     }

//     fmt.Fprintf(w, "Messages:\n")
//     for _, m := range rList.Messages {
//         fmt.Fprintf(w, "Message ID: %s\n", m.Id)
//     }
// }


// func main() {
//     goth.UseProviders(
//         facebook.New("FACEBOOK_CLIENT_ID", "FACEBOOK_CLIENT_SECRET", "http://localhost:8080/auth/facebook/callback"),
//         google.New("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "http://localhost:8080/auth/google/callback", "email", "profile"),
//     )

//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         http.ServeFile(w, r, "index.html")
//     })

//     http.HandleFunc("/auth/facebook", func(w http.ResponseWriter, r *http.Request) {
//         gothic.BeginAuthHandler(w, r)
//     })

//     http.HandleFunc("/auth/facebook/callback", func(w http.ResponseWriter, r *http.Request) {
//         user, err := gothic.CompleteUserAuth(w, r)
//         if err != nil {
//             fmt.Fprintln(w, "Login failed:", err)
//             return
//         }
//         fmt.Fprintf(w, "✅ Facebook Login Successful!\n\nName: %s\nEmail: %s\n", user.Name, user.Email)
//     })

//     http.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
//         gothic.BeginAuthHandler(w, r)
//     })

//     http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
//         user, err := gothic.CompleteUserAuth(w, r)
//         if err != nil {
//             fmt.Fprintln(w, "Login failed:", err)
//             return
//         }
//         fmt.Fprintf(w, "✅ Google Login Successful!\n\nName: %s\nEmail: %s\n", user.Name, user.Email)
//     })

//     fmt.Println("Server running at http://localhost:8080")
//     log.Fatal(http.ListenAndServe(":8080", nil))
// }


// func handleEmailSubscription(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 					return
// 	}

// 	email := r.FormValue("email")
// 	if email == "" {
// 					http.Error(w, "Email is required", http.StatusBadRequest)
// 					return
// 	}

// 	// Save email to a file
// 	file, err := os.OpenFile("emails.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 					http.Error(w, "Failed to save email", http.StatusInternalServerError)
// 					return
// 	}
// 	defer file.Close()

// 	if _, err := file.WriteString(email + "\n"); err != nil {
// 					http.Error(w, "Failed to write to file", http.StatusInternalServerError)
// 					return
// 	}

// 	// Send confirmation email
// 	go sendConfirmationEmail(email)

// 	fmt.Fprintf(w, "✅ Thanks for subscribing! A confirmation has been sent to %s", email)
// }


// func main() {
//     http.HandleFunc("/", serveHome)
//     http.HandleFunc("/subscribe/email", handleEmailSubscription)
//     http.HandleFunc("/subscribe/facebook", simulateFacebook)
//     http.HandleFunc("/subscribe/google", simulateGoogle)

//     fmt.Println("Server running at http://localhost:8080")
//     log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func serveHome(w http.ResponseWriter, r *http.Request) {
//     http.ServeFile(w, r, "subscribe.html")
// }

// func handleEmailSubscription(w http.ResponseWriter, r *http.Request) {
//     if r.Method != http.MethodPost {
//         http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//         return
//     }

//     email := r.FormValue("email")
//     if email == "" {
//         http.Error(w, "Email is required", http.StatusBadRequest)
//         return
//     }

//     // In production, you would store this email in a DB or send it to a mailing service
//     fmt.Fprintf(w, "✅ Thanks for subscribing with: %s", email)
// }

// func simulateFacebook(w http.ResponseWriter, r *http.Request) {
//     fmt.Fprint(w, "🔵 Facebook subscription would go here.")
// }

// func simulateGoogle(w http.ResponseWriter, r *http.Request) {
//     fmt.Fprint(w, "🟢 Google subscription would go here.")
// }


// var db *sql.DB

// func main() {
//     var err error
//     db, err = sql.Open("sqlite3", "./subscribers.db")
//     if err != nil {
//         log.Fatal(err)
//     }
//     defer db.Close()

//     createTable()

//     http.HandleFunc("/", serveHome)
//     http.HandleFunc("/subscribe/email", handleEmailSubscription)
//     http.HandleFunc("/subscribe/facebook", simulateFacebook)
//     http.HandleFunc("/subscribe/google", simulateGoogle)

//     fmt.Println("Server running at http://localhost:8080")
//     log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func createTable() {
//     query := `
//     CREATE TABLE IF NOT EXISTS subscribers (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         email TEXT UNIQUE NOT NULL
//     );
//     `
//     _, err := db.Exec(query)
//     if err != nil {
//         log.Fatalf("❌ Failed to create table: %v", err)
//     }
// }

// func handleEmailSubscription(w http.ResponseWriter, r *http.Request) {
//      email := r.FormValue("email")
//     if email == "" {
//         http.Error(w, "Email is required", http.StatusBadRequest)
//         return
//     }

//     // Save email to the database
//     _, err := db.Exec("INSERT OR IGNORE INTO subscribers (email) VALUES (?)", email)
//     if err != nil {
//         http.Error(w, "Failed to save email", http.StatusInternalServerError)
//         log.Println("DB error:", err)
//         return
//     }

//     go sendConfirmationEmail(email) // fire-and-forget

//     fmt.Fprintf(w, "✅ Subscription successful,thanks for subscribing! A confirmation has been sent to %s", email)



// func sendConfirmationEmail(to, link string) {
//     from := "yourgmail@gmail.com"
//     password := "your-app-password"
//     subject := "Verify your subscription"
//     body := fmt.Sprintf("Click the link to verify your subscription:\n\n%s", link)


//     msg := "From: " + from + "\n" +
//         "To: " + to + "\n" +
//     "subject: " + subject + "\n\n" + body

//     err := smtp.SendMail("smtp.gmail.com:587",
//         smtp.PlainAuth("", from, password, "smtp.gmail.com"),
//         from, []string{to}, []byte(msg))

//     if err != nil {
//         log.Println("❌ Failed to send email to %s: %v", to, err)
//     } else {
//         log.Printf("✅ Confirmation email sent to %s", to)
//     }
// }
