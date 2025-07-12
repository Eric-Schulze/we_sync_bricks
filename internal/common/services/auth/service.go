package auth

import (
	"errors"
	"html/template"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

// AuthRepositoryInterface defines the interface for auth repository operations
type AuthRepositoryInterface interface {
	CreateUser(req *RegisterRequest) (*models.User, error)
	GetUserByEmailOrUsername(login string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	UpdateUserLastLogin(userID int64) error
}

type AuthService struct {
	repo      AuthRepositoryInterface
	jwtSecret []byte
}

// NewAuthService creates a new auth service
func NewAuthService(repo AuthRepositoryInterface, jwtSecret []byte) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Login authenticates a user with email or username and password
func (s *AuthService) Login(login, password string) (*LoginResponse, error) {
	if login == "" || password == "" {
		return nil, errors.New("login and password are required")
	}

	// Get user from database
	user, err := s.repo.GetUserByEmailOrUsername(login)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Verify password hash
	if !s.verifyPassword(user.Password, password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Update last login
	s.repo.UpdateUserLastLogin(user.ID)

	return &LoginResponse{
		Token: token,
		User:  "user_data_placeholder", // TODO: Return actual user data
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	if req.Email == "" || req.Password == "" || req.Username == "" {
		return nil, errors.New("email, password, and username are required")
	}

	// Validate password strength
	if err := s.validatePassword(req.Password); err != nil {
		return nil, errors.New("password does not meet requirements: " + err.Error())
	}

	// Hash password before storing
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	req.Password = hashedPassword

	// Create user in database
	user, err := s.repo.CreateUser(req)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims := token.Claims.(*Claims)
	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// generateToken creates a JWT token for the user
func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// hashPassword hashes a password using bcrypt
func (s *AuthService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// verifyPassword compares a hashed password with a plain text password
func (s *AuthService) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// validatePassword validates password strength and requirements
func (s *AuthService) validatePassword(password string) error {
	// Minimum entropy of 60 bits (good security balance)
	// This typically requires 8+ characters with mixed case, numbers, symbols
	const minEntropyBits = 60

	return passwordvalidator.Validate(password, minEntropyBits)
}

// InitializeAuthHandler creates a fully initialized auth handler with all dependencies
func InitializeAuthHandler(dbService models.DBService, templates *template.Template, jwtSecret []byte) *AuthHandler {
	repo := NewAuthRepository(dbService)
	service := NewAuthService(repo, jwtSecret)
	handler := NewAuthHandler(service, templates)
	return handler
}
