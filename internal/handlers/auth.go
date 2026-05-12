package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"go-news-site-production/internal/database"
	"go-news-site-production/internal/middleware"
	"go-news-site-production/internal/models"
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	pass_1 := r.FormValue("pass_1")
	pass_2 := r.FormValue("pass_2")

	t, err := template.ParseFiles("templates/registration.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if name == "" || email == "" || pass_1 == "" {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormName    string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/registration",
			FormName:    name,
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "Please fill all registration fields",
		}
		t.ExecuteTemplate(w, "registration", data)
		return
	}

	if pass_1 != pass_2 {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormName    string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/registration",
			FormName:    name,
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "Passwords do not match",
		}
		t.ExecuteTemplate(w, "registration", data)
		return
	}

	hashedPassword, err := HashPassword(pass_1)
	if err != nil {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormName    string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/registration",
			FormName:    name,
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "Error creating password hash",
		}
		t.ExecuteTemplate(w, "registration", data)
		return
	}

	query := fmt.Sprintf("INSERT INTO users (name, email, password) VALUES(%s, %s, %s)",
		database.GetPlaceholder(0), database.GetPlaceholder(1), database.GetPlaceholder(2))
	_, err = database.DB.Exec(query, name, email, hashedPassword)
	if err != nil {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormName    string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/registration",
			FormName:    name,
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "User with this email already exists",
		}
		t.ExecuteTemplate(w, "registration", data)
		return
	}

	SetFlashMessage(w, "success", "Registration successful. Please login")
	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func AuthUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	pass := r.FormValue("pass")

	t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html", "templates/partials/flash.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if email == "" || pass == "" {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/auth",
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "Please fill all authorization fields",
		}
		t.ExecuteTemplate(w, "auth", data)
		return
	}

	var user models.User
	var hashedPassword string

	query := fmt.Sprintf("SELECT id, password FROM users WHERE email = %s LIMIT 1", database.GetPlaceholder(0))
	err = database.DB.QueryRow(query, email).Scan(&user.Id, &hashedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			data := struct {
				IsAuth      bool
				CurrentPath string
				FormEmail   string
				FlashType   string
				FlashMsg    string
			}{
				IsAuth:      false,
				CurrentPath: "/auth",
				FormEmail:   email,
				FlashType:   "error",
				FlashMsg:    "Invalid email or password",
			}
			t.ExecuteTemplate(w, "auth", data)
			return
		} else {
			http.Error(w, "Database error: Unable to process login", http.StatusInternalServerError)
			return
		}
	}

	if !CheckPasswordHash(pass, hashedPassword) {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/auth",
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "Invalid email or password",
		}
		t.ExecuteTemplate(w, "auth", data)
		return
	}

	sessionToken, err := middleware.CreateSession(user.Id)
	if err != nil {
		data := struct {
			IsAuth      bool
			CurrentPath string
			FormEmail   string
			FlashType   string
			FlashMsg    string
		}{
			IsAuth:      false,
			CurrentPath: "/auth",
			FormEmail:   email,
			FlashType:   "error",
			FlashMsg:    "Error: Session creation failed",
		}
		t.ExecuteTemplate(w, "auth", data)
		return
	}

	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	SetFlashMessage(w, "success", "You have successfully logged in")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	middleware.DeleteSession(r)

	cookie := http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	SetFlashMessage(w, "success", "You have been logged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}