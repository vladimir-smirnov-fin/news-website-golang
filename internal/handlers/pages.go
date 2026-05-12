package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"go-news-site-production/internal/database"
	"go-news-site-production/internal/middleware"
	"go-news-site-production/internal/models"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := database.DB.Query("SELECT id, title, anons, full_text, user_id FROM articles ORDER BY id DESC")
	if err != nil {
		http.Error(w, "Database error: Unable to fetch articles", http.StatusInternalServerError)
		return
	}
	defer res.Close()

	posts := []models.ArticleWithAuthor{}
	for res.Next() {
		var post models.Article
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText, &post.UserId)
		if err != nil {
			http.Error(w, "Database error: Unable to read article data", http.StatusInternalServerError)
			return
		}

		var authorName string
		query := fmt.Sprintf("SELECT name FROM users WHERE id = %s", database.GetPlaceholder(0))
		err = database.DB.QueryRow(query, post.UserId).Scan(&authorName)
		if err != nil {
			authorName = "Unknown author"
		}

		posts = append(posts, models.ArticleWithAuthor{
			Article:    post,
			AuthorName: authorName,
		})
	}

	userName := ""
	if userId, isAuth := middleware.GetCurrentUserId(r); isAuth {
		query := fmt.Sprintf("SELECT name FROM users WHERE id = %s", database.GetPlaceholder(0))
		err = database.DB.QueryRow(query, userId).Scan(&userName)
		if err != nil {
			userName = "User"
		}
	}

	flashType, flashMsg := GetFlashMessage(r)
	ClearFlashMessage(w)

	data := struct {
		Articles    []models.ArticleWithAuthor
		IsAuth      bool
		CurrentPath string
		UserName    string
		FlashType   string
		FlashMsg    string
	}{
		Articles:    posts,
		IsAuth:      middleware.IsAuthenticated(r),
		CurrentPath: r.URL.Path,
		UserName:    userName,
		FlashType:   flashType,
		FlashMsg:    flashMsg,
	}

	t.ExecuteTemplate(w, "index", data)
}

func Contacts(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/contacts.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		IsAuth      bool
		CurrentPath string
	}{
		IsAuth:      middleware.IsAuthenticated(r),
		CurrentPath: r.URL.Path,
	}

	t.ExecuteTemplate(w, "contacts", data)
}

func Registration(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/registration.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		IsAuth      bool
		CurrentPath string
		FormName    string
		FormEmail   string
		FlashType   string
		FlashMsg    string
	}{
		IsAuth:      middleware.IsAuthenticated(r),
		CurrentPath: r.URL.Path,
		FormName:    "",
		FormEmail:   "",
		FlashType:   "",
		FlashMsg:    "",
	}

	err = t.ExecuteTemplate(w, "registration", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Auth(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashType, flashMsg := GetFlashMessage(r)
	ClearFlashMessage(w)

	data := struct {
		IsAuth      bool
		CurrentPath string
		FormEmail   string
		FlashType   string
		FlashMsg    string
	}{
		IsAuth:      middleware.IsAuthenticated(r),
		CurrentPath: r.URL.Path,
		FormEmail:   "",
		FlashType:   flashType,
		FlashMsg:    flashMsg,
	}

	err = t.ExecuteTemplate(w, "auth", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}