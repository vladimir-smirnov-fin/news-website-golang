package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"go-news-site-production/internal/database"
	"go-news-site-production/internal/middleware"
	"go-news-site-production/internal/models"
)

func CreateEditForm(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("templates/create-edit.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashType, flashMsg := GetFlashMessage(r)
	ClearFlashMessage(w)

	data := struct {
		Article     models.Article
		IsAuth      bool
		CurrentPath string
		FlashType   string
		FlashMsg    string
		FormTitle   string
		FormAnons   string
		FormText    string
	}{
		Article:     models.Article{},
		IsAuth:      true,
		CurrentPath: r.URL.Path,
		FlashType:   flashType,
		FlashMsg:    flashMsg,
		FormTitle:   "",
		FormAnons:   "",
		FormText:    "",
	}

	t.ExecuteTemplate(w, "create-edit", data)
}

func CreateArticle(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	t, err := template.ParseFiles("templates/create-edit.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if title == "" || anons == "" || full_text == "" {
		data := struct {
			Article     models.Article
			IsAuth      bool
			CurrentPath string
			FlashType   string
			FlashMsg    string
			FormTitle   string
			FormAnons   string
			FormText    string
		}{
			Article:     models.Article{},
			IsAuth:      true,
			CurrentPath: "/create",
			FlashType:   "error",
			FlashMsg:    "Please fill all fields - title, anons and text",
			FormTitle:   title,
			FormAnons:   anons,
			FormText:    full_text,
		}
		t.ExecuteTemplate(w, "create-edit", data)
		return
	}

	userId, ok := middleware.GetCurrentUserId(r)
	if !ok {
		data := struct {
			Article     models.Article
			IsAuth      bool
			CurrentPath string
			FlashType   string
			FlashMsg    string
			FormTitle   string
			FormAnons   string
			FormText    string
		}{
			Article:     models.Article{},
			IsAuth:      true,
			CurrentPath: "/create",
			FlashType:   "error",
			FlashMsg:    "Error: Authorization failed. Please login again",
			FormTitle:   title,
			FormAnons:   anons,
			FormText:    full_text,
		}
		t.ExecuteTemplate(w, "create-edit", data)
		return
	}

	query := fmt.Sprintf("INSERT INTO articles (title, anons, full_text, user_id) VALUES(%s, %s, %s, %s)",
		database.GetPlaceholder(0), database.GetPlaceholder(1), database.GetPlaceholder(2), database.GetPlaceholder(3))

	_, err = database.DB.Exec(query, title, anons, full_text, userId)
	if err != nil {
		http.Error(w, "Database error: Unable to save article", http.StatusInternalServerError)
		return
	}

	SetFlashMessage(w, "success", "Article created successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ShowArticle(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	idStr := parts[2]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("templates/show.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var post models.Article
	query := fmt.Sprintf("SELECT id, title, anons, full_text, user_id FROM articles WHERE id = %s", database.GetPlaceholder(0))
	err = database.DB.QueryRow(query, id).Scan(&post.Id, &post.Title, &post.Anons, &post.FullText, &post.UserId)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Article not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: Unable to fetch article", http.StatusInternalServerError)
		}
		return
	}

	var authorName string
	query2 := fmt.Sprintf("SELECT name FROM users WHERE id = %s", database.GetPlaceholder(0))
	err = database.DB.QueryRow(query2, post.UserId).Scan(&authorName)
	if err != nil {
		authorName = "Unknown author"
	}

	currentUserId, isAuth := middleware.GetCurrentUserId(r)
	canEdit := isAuth && middleware.CanEditArticle(currentUserId, post.UserId)

	flashType, flashMsg := GetFlashMessage(r)
	ClearFlashMessage(w)

	data := struct {
		Article     models.Article
		AuthorName  string
		IsAuth      bool
		CanEdit     bool
		CurrentPath string
		FlashType   string
		FlashMsg    string
	}{
		Article:     post,
		AuthorName:  authorName,
		IsAuth:      isAuth,
		CanEdit:     canEdit,
		CurrentPath: r.URL.Path,
		FlashType:   flashType,
		FlashMsg:    flashMsg,
	}

	t.ExecuteTemplate(w, "show", data)
}

func EditArticleForm(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	idStr := parts[2]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	t, err := template.ParseFiles("templates/create-edit.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var post models.Article
	query := fmt.Sprintf("SELECT id, title, anons, full_text, user_id FROM articles WHERE id = %s", database.GetPlaceholder(0))
	err = database.DB.QueryRow(query, id).Scan(&post.Id, &post.Title, &post.Anons, &post.FullText, &post.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Article not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: Unable to fetch article", http.StatusInternalServerError)
		}
		return
	}

	currentUserId, isAuth := middleware.GetCurrentUserId(r)
	if !isAuth || currentUserId != post.UserId {
		http.Error(w, "You don't have permission to edit this article", http.StatusForbidden)
		return
	}

	flashType, flashMsg := GetFlashMessage(r)
	ClearFlashMessage(w)

	data := struct {
		Article     models.Article
		IsAuth      bool
		CurrentPath string
		FlashType   string
		FlashMsg    string
	}{
		Article:     post,
		IsAuth:      middleware.IsAuthenticated(r),
		CurrentPath: r.URL.Path,
		FlashType:   flashType,
		FlashMsg:    flashMsg,
	}

	err = t.ExecuteTemplate(w, "create-edit", data)
	if err != nil {
		http.Error(w, "Error rendering: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	idStr := parts[2]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	if title == "" || anons == "" || full_text == "" {
		SetFlashMessage(w, "error", "Please fill all fields")
		http.Redirect(w, r, "/article/"+idStr+"/edit", http.StatusSeeOther)
		return
	}

	var articleUserId int
	query1 := fmt.Sprintf("SELECT user_id FROM articles WHERE id = %s", database.GetPlaceholder(0))
	err = database.DB.QueryRow(query1, id).Scan(&articleUserId)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	currentUserId, isAuth := middleware.GetCurrentUserId(r)
	if !isAuth || currentUserId != articleUserId {
		http.Error(w, "You don't have permission to edit this article", http.StatusForbidden)
		return
	}

	query2 := fmt.Sprintf("UPDATE articles SET title = %s, anons = %s, full_text = %s WHERE id = %s",
		database.GetPlaceholder(0), database.GetPlaceholder(1), database.GetPlaceholder(2), database.GetPlaceholder(3))
	_, err = database.DB.Exec(query2, title, anons, full_text, id)
	if err != nil {
		http.Error(w, "Database error: Unable to update article", http.StatusInternalServerError)
		return
	}

	SetFlashMessage(w, "success", "Article updated successfully")
	http.Redirect(w, r, "/article/"+idStr, http.StatusSeeOther)
}

func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !middleware.IsAuthenticated(r) {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	idStr := parts[2]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	var articleUserId int
	query1 := fmt.Sprintf("SELECT user_id FROM articles WHERE id = %s", database.GetPlaceholder(0))
	err = database.DB.QueryRow(query1, id).Scan(&articleUserId)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	currentUserId, isAuth := middleware.GetCurrentUserId(r)
	if !isAuth || currentUserId != articleUserId {
		http.Error(w, "You don't have permission to delete this article", http.StatusForbidden)
		return
	}

	query2 := fmt.Sprintf("DELETE FROM articles WHERE id = %s", database.GetPlaceholder(0))
	_, err = database.DB.Exec(query2, id)
	if err != nil {
		http.Error(w, "Database error: Unable to delete article", http.StatusInternalServerError)
		return
	}

	SetFlashMessage(w, "success", "Article deleted successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}