package auth

import (
	"html/template"
	"net/http"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type AuthHandler struct {
	service   *AuthService
	templates *template.Template
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service *AuthService, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		service:   service,
		templates: templates,
	}
}

// HandleLoginPage displays the login page
func (h *AuthHandler) HandleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title":       "Login",
			"CurrentPage": "login",
		}

		logger.Info("Handling request to Login page")

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := h.templates.ExecuteTemplate(w, "login-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If not an HTMX request, render the full page
		if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
			logger.Error("Error handling request to Login page", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandleLogin processes the login form submission
func (h *AuthHandler) HandleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		login := r.Form.Get("login")
		password := r.Form.Get("password")

		if login == "" || password == "" {
			http.Error(w, "Login and password are required", http.StatusBadRequest)
			return
		}

		loginResp, err := h.service.Login(login, password)
		if err != nil {
			logger.Error("Login failed", "login", login, "error", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Set JWT token as HTTP-only cookie
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    loginResp.Token,
			Path:     "/",
			MaxAge:   24 * 60 * 60, // 24 hours
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		logger.Info("User logged in successfully", "login", login)

		// Redirect to dashboard or return success response
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/")
			w.WriteHeader(http.StatusOK)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

// HandleRegisterPage displays the registration page
func (h *AuthHandler) HandleRegisterPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Title":       "Register",
			"CurrentPage": "register",
		}

		logger.Info("Handling request to Register page")

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := h.templates.ExecuteTemplate(w, "register-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If not an HTMX request, render the full page
		if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
			logger.Error("Error handling request to Register page", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandleRegister processes the registration form submission
func (h *AuthHandler) HandleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		req := &RegisterRequest{
			Username:  r.Form.Get("username"),
			Email:     r.Form.Get("email"),
			Password:  r.Form.Get("password"),
			FirstName: r.Form.Get("first_name"),
			LastName:  r.Form.Get("last_name"),
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
			return
		}

		user, err := h.service.Register(req)
		if err != nil {
			logger.Error("Registration failed", "email", req.Email, "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("User registered successfully", "email", req.Email, "user_id", user.ID)

		// Redirect to login page or auto-login
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/auth/login")
			w.WriteHeader(http.StatusOK)
		} else {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		}
	}
}

// HandleLogout processes user logout
func (h *AuthHandler) HandleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear the auth token cookie
		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // Delete cookie
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		logger.Info("User logged out")

		// Redirect to login page
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/auth/login")
			w.WriteHeader(http.StatusOK)
		} else {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		}
	}
}