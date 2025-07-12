package profile

import (
	"html/template"
	"net/http"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type ProfileHandler struct {
	service   *ProfileService
	templates *template.Template
	jwtSecret []byte
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(service *ProfileService, templates *template.Template, jwtSecret []byte) *ProfileHandler {
	return &ProfileHandler{
		service:   service,
		templates: templates,
		jwtSecret: jwtSecret,
	}
}

// HandleProfilePage displays the user's profile page
func (h *ProfileHandler) HandleProfilePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// Get full profile data with credentials
		profile, err := h.service.GetProfileWithCredentials(user.ID)
		if err != nil {
			logger.Error("Error getting user profile", "user_id", user.ID, "error", err)
			http.Error(w, "Error loading profile", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title":       "Profile",
			"CurrentPage": "profile",
			"Profile":     profile,
			"User":        profile.User, // For template compatibility
		}

		logger.Info("Handling request to Profile page", "user_id", user.ID)

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := h.templates.ExecuteTemplate(w, "profile-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// If not an HTMX request, render the full page
		if err := h.templates.ExecuteTemplate(w, "profile.html", data); err != nil {
			logger.Error("Error handling request to Profile page", "user_id", user.ID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandleEditProfilePage displays the profile edit form
func (h *ProfileHandler) HandleEditProfilePage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// Get full profile data with credentials
		profile, err := h.service.GetProfileWithCredentials(user.ID)
		if err != nil {
			logger.Error("Error getting user profile for edit", "user_id", user.ID, "error", err)
			http.Error(w, "Error loading profile", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Title":       "Edit Profile",
			"CurrentPage": "profile",
			"Profile":     profile,
			"User":        profile.User, // For template compatibility
		}

		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			if err := h.templates.ExecuteTemplate(w, "profile-edit-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Render the edit form template
		if err := h.templates.ExecuteTemplate(w, "profile-edit.html", data); err != nil {
			logger.Error("Error handling request to Profile edit page", "user_id", user.ID, "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// HandleUpdateProfile processes the profile update form submission
func (h *ProfileHandler) HandleUpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		req := &ProfileUpdateRequest{
			FirstName: r.Form.Get("first_name"),
			LastName:  r.Form.Get("last_name"),
			Email:     r.Form.Get("email"),
			Username:  r.Form.Get("username"),
		}

		updatedUser, err := h.service.UpdateProfile(user.ID, req)
		if err != nil {
			logger.Error("Profile update failed", "user_id", user.ID, "error", err)
			http.Error(w, "Profile update failed", http.StatusInternalServerError)
			return
		}

		logger.Info("Profile updated successfully", "user_id", user.ID)

		// Return updated profile data
		data := map[string]interface{}{
			"User":    updatedUser,
			"Success": true,
			"Message": "Profile updated successfully",
		}

		// Redirect or return success response
		if r.Header.Get("HX-Request") == "true" {
			// Return updated profile content for HTMX
			if err := h.templates.ExecuteTemplate(w, "profile-content", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
		}
	}
}

// HandleChangePassword processes password change requests
func (h *ProfileHandler) HandleChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			logger.Error("Error parsing password change form", "user_id", user.ID, "error", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		req := &PasswordChangeRequest{
			CurrentPassword: r.Form.Get("current_password"),
			NewPassword:     r.Form.Get("new_password"),
			ConfirmPassword: r.Form.Get("confirm_password"),
		}

		err := h.service.ChangePassword(user.ID, req)
		if err != nil {
			logger.Error("Password change failed", "user_id", user.ID, "error", err)

			// Return error response for HTMX
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`<div class="error-message" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px; margin: 10px 0;">Error: ` + err.Error() + `</div>`))
				return
			}

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("Password changed successfully", "user_id", user.ID)

		// Return success response
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="success-message" style="color: green; padding: 10px; border: 1px solid green; border-radius: 4px; margin: 10px 0;">Password changed successfully!</div>`))
		} else {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
		}
	}
}

// HandleUpdateAPIKeys processes API key update requests
func (h *ProfileHandler) HandleUpdateAPIKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			logger.Error("Error parsing API key form", "user_id", user.ID, "error", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		provider := r.Form.Get("provider")
		req := &APIKeyUpdateRequest{
			Provider:       provider,
			ConsumerKey:    r.Form.Get("consumer_key"),
			ConsumerSecret: r.Form.Get("consumer_secret"),
			Token:          r.Form.Get("token"),
			TokenSecret:    r.Form.Get("token_secret"),
		}

		err := h.service.UpdateAPICredentials(user.ID, req)
		if err != nil {
			logger.Error("API key update failed", "user_id", user.ID, "provider", provider, "error", err)

			// Return error response for HTMX
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`<div class="error-message" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px; margin: 10px 0;">Error: ` + err.Error() + `</div>`))
				return
			}

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("API credentials updated successfully", "user_id", user.ID, "provider", provider)

		// Return success response
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="success-message" style="color: green; padding: 10px; border: 1px solid green; border-radius: 4px; margin: 10px 0;">` + provider + ` API credentials updated successfully!</div>`))
		} else {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
		}
	}
}

// HandleDeleteAPIKeys processes API key deletion requests
func (h *ProfileHandler) HandleDeleteAPIKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			logger.Error("Error parsing API key deletion form", "user_id", user.ID, "error", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		provider := r.Form.Get("provider")

		err := h.service.DeleteAPICredentials(user.ID, provider)
		if err != nil {
			logger.Error("API key deletion failed", "user_id", user.ID, "provider", provider, "error", err)

			// Return error response for HTMX
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`<div class="error-message" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px; margin: 10px 0;">Error: ` + err.Error() + `</div>`))
				return
			}

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("API credentials deleted successfully", "user_id", user.ID, "provider", provider)

		// Return success response
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div class="success-message" style="color: green; padding: 10px; border: 1px solid green; border-radius: 4px; margin: 10px 0;">` + provider + ` API credentials deleted successfully!</div>`))
		} else {
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
		}
	}
}
