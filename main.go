package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"go-news-site-production/internal/database"
	"go-news-site-production/internal/handlers"
	"go-news-site-production/internal/middleware"
)

func main() {
	// Initialize database connection
	if err := database.InitDB(); err != nil {
		panic(err)
	}

	// Add this right after database.InitDB() in main.go:
	if err := database.MigrateDB(); err != nil {
		panic(err)
	}
	
	defer database.CloseDB()

	// Start session cleanup goroutine
	go middleware.CleanupExpiredSessions()

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Page routes
	http.HandleFunc("/", handlers.Index)
	http.HandleFunc("/contacts", handlers.Contacts)
	http.HandleFunc("/create", handlers.CreateEditForm)
	http.HandleFunc("/create-article", handlers.CreateArticle)
	http.HandleFunc("/registration", handlers.Registration)
	http.HandleFunc("/add-user", handlers.AddUser)
	http.HandleFunc("/auth", handlers.Auth)
	http.HandleFunc("/auth-user", handlers.AuthUser)
	http.HandleFunc("/logout", handlers.Logout)

	// Article routes with dynamic paths
	http.HandleFunc("/article/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		if strings.HasSuffix(path, "/edit") && method == "GET" {
			handlers.EditArticleForm(w, r)
		} else if strings.HasSuffix(path, "/edit") && method == "POST" {
			handlers.UpdateArticle(w, r)
		} else if strings.HasSuffix(path, "/delete") && method == "POST" {
			handlers.DeleteArticle(w, r)
		} else if method == "GET" {
			handlers.ShowArticle(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Get port from environment variable (Render sets this automatically)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on http://localhost:%s\n", port)
	fmt.Println("Press Ctrl+C to stop the server")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(fmt.Sprintf("Server failed to start: %v", err))
	}
}