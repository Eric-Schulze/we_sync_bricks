package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"golang.org/x/crypto/bcrypt"
)

// MockAuthRepository implements the auth repository interface for testing
type MockAuthRepository struct {
	users              map[string]*models.User // key: email or username
	usersByID          map[int64]*models.User
	createUserErr      error
	getUserErr         error
	updateLastLoginErr error
	nextUserID         int64
}

func NewMockAuthRepository() *MockAuthRepository {
	return &MockAuthRepository{
		users:      make(map[string]*models.User),
		usersByID:  make(map[int64]*models.User),
		nextUserID: 1,
	}
}

func (m *MockAuthRepository) CreateUser(req *RegisterRequest) (*models.User, error) {
	if m.createUserErr != nil {
		return nil, m.createUserErr
	}

	// Check if user already exists
	if _, exists := m.users[req.Email]; exists {
		return nil, errors.New("user with email already exists")
	}
	if _, exists := m.users[req.Username]; exists {
		return nil, errors.New("user with username already exists")
	}

	user := &models.User{
		ID:       m.nextUserID,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		CreatedAt: time.Now(),
	}

	if req.FirstName != "" {
		user.FirstName = &req.FirstName
	}
	if req.LastName != "" {
		user.LastName = &req.LastName
	}

	m.users[req.Email] = user
	m.users[req.Username] = user
	m.usersByID[m.nextUserID] = user
	m.nextUserID++

	return user, nil
}

func (m *MockAuthRepository) GetUserByEmailOrUsername(login string) (*models.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}

	user, exists := m.users[login]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (m *MockAuthRepository) GetUserByID(id int64) (*models.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}

	user, exists := m.usersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (m *MockAuthRepository) UpdateUserLastLogin(userID int64) error {
	return m.updateLastLoginErr
}

func (m *MockAuthRepository) AddUser(user *models.User) {
	m.users[user.Email] = user
	m.users[user.Username] = user
	m.usersByID[user.ID] = user
}

func (m *MockAuthRepository) SetCreateUserError(err error) {
	m.createUserErr = err
}

func (m *MockAuthRepository) SetGetUserError(err error) {
	m.getUserErr = err
}

func (m *MockAuthRepository) SetUpdateLastLoginError(err error) {
	m.updateLastLoginErr = err
}

func TestNewAuthService(t *testing.T) {
	repo := NewMockAuthRepository()
	jwtSecret := []byte("test-secret")

	service := NewAuthService(repo, jwtSecret)

	if service.repo == nil {
		t.Error("Expected repository to be set")
	}
	if string(service.jwtSecret) != string(jwtSecret) {
		t.Error("Expected JWT secret to be set correctly")
	}
}

func TestAuthService_hashPassword(t *testing.T) {
	service := NewAuthService(NewMockAuthRepository(), []byte("test-secret"))

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "test123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt can hash empty strings
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-that-should-still-work-fine",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.hashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("hashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify it's a valid bcrypt hash
				if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password)); err != nil {
					t.Errorf("Generated hash does not match original password: %v", err)
				}
			}
		})
	}
}

func TestAuthService_verifyPassword(t *testing.T) {
	service := NewAuthService(NewMockAuthRepository(), []byte("test-secret"))

	// Generate a known hash for testing
	password := "test123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate hash for testing: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		want           bool
	}{
		{
			name:           "correct password",
			hashedPassword: string(hash),
			password:       password,
			want:           true,
		},
		{
			name:           "incorrect password",
			hashedPassword: string(hash),
			password:       "wrong",
			want:           false,
		},
		{
			name:           "empty password against hash",
			hashedPassword: string(hash),
			password:       "",
			want:           false,
		},
		{
			name:           "empty hash",
			hashedPassword: "",
			password:       password,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.verifyPassword(tt.hashedPassword, tt.password)
			if got != tt.want {
				t.Errorf("verifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthService_validatePassword(t *testing.T) {
	service := NewAuthService(NewMockAuthRepository(), []byte("test-secret"))

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "weak password",
			password: "weak",
			wantErr:  true,
		},
		{
			name:     "strong password",
			password: "StrongPass123!",
			wantErr:  false,
		},
		{
			name:     "medium password",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "another strong password",
			password: "MySecure#Pass2024",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name    string
		req     *RegisterRequest
		setup   func(*MockAuthRepository)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful registration",
			req: &RegisterRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "StrongPass123!",
				FirstName: "Test",
				LastName:  "User",
			},
			setup:   func(repo *MockAuthRepository) {},
			wantErr: false,
		},
		{
			name: "missing email",
			req: &RegisterRequest{
				Username: "testuser",
				Password: "StrongPass123!",
			},
			setup:   func(repo *MockAuthRepository) {},
			wantErr: true,
			errMsg:  "email, password, and username are required",
		},
		{
			name: "missing password",
			req: &RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
			},
			setup:   func(repo *MockAuthRepository) {},
			wantErr: true,
			errMsg:  "email, password, and username are required",
		},
		{
			name: "missing username",
			req: &RegisterRequest{
				Email:    "test@example.com",
				Password: "StrongPass123!",
			},
			setup:   func(repo *MockAuthRepository) {},
			wantErr: true,
			errMsg:  "email, password, and username are required",
		},
		{
			name: "weak password",
			req: &RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
			setup:   func(repo *MockAuthRepository) {},
			wantErr: true,
			errMsg:  "password does not meet requirements",
		},
		{
			name: "repository error",
			req: &RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "StrongPass123!",
			},
			setup: func(repo *MockAuthRepository) {
				repo.SetCreateUserError(errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockAuthRepository()
			tt.setup(repo)
			service := NewAuthService(repo, []byte("test-secret"))

			// Store original password for verification
			originalPassword := tt.req.Password

			user, err := service.Register(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil || (tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg) {
					t.Errorf("Register() error = %v, expected error containing %v", err, tt.errMsg)
				}
				return
			}

			if user == nil {
				t.Error("Register() returned nil user on success")
				return
			}

			// Verify password was hashed (should not equal original)
			if user.Password == originalPassword {
				t.Error("Password was not hashed")
			}

			// Verify hash is valid against original password
			if !service.verifyPassword(user.Password, originalPassword) {
				t.Error("Password hash verification failed")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		setup    func(*MockAuthRepository)
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "successful login with email",
			login:    "test@example.com",
			password: "StrongPass123!",
			setup: func(repo *MockAuthRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Password: string(hash),
					CreatedAt: time.Now(),
				}
				repo.AddUser(user)
			},
			wantErr: false,
		},
		{
			name:     "successful login with username",
			login:    "testuser",
			password: "StrongPass123!",
			setup: func(repo *MockAuthRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Password: string(hash),
					CreatedAt: time.Now(),
				}
				repo.AddUser(user)
			},
			wantErr: false,
		},
		{
			name:     "empty login",
			login:    "",
			password: "StrongPass123!",
			setup:    func(repo *MockAuthRepository) {},
			wantErr:  true,
			errMsg:   "login and password are required",
		},
		{
			name:     "empty password",
			login:    "test@example.com",
			password: "",
			setup:    func(repo *MockAuthRepository) {},
			wantErr:  true,
			errMsg:   "login and password are required",
		},
		{
			name:     "user not found",
			login:    "nonexistent@example.com",
			password: "StrongPass123!",
			setup:    func(repo *MockAuthRepository) {},
			wantErr:  true,
			errMsg:   "user not found",
		},
		{
			name:     "incorrect password",
			login:    "test@example.com",
			password: "WrongPassword123!",
			setup: func(repo *MockAuthRepository) {
				hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Password: string(hash),
					CreatedAt: time.Now(),
				}
				repo.AddUser(user)
			},
			wantErr: true,
			errMsg:  "invalid credentials",
		},
		{
			name:     "repository error",
			login:    "test@example.com",
			password: "StrongPass123!",
			setup: func(repo *MockAuthRepository) {
				repo.SetGetUserError(errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockAuthRepository()
			tt.setup(repo)
			service := NewAuthService(repo, []byte("test-secret"))

			resp, err := service.Login(tt.login, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil || (tt.errMsg != "" && err.Error() != tt.errMsg) {
					t.Errorf("Login() error = %v, expected error %v", err, tt.errMsg)
				}
				return
			}

			if resp == nil {
				t.Error("Login() returned nil response on success")
				return
			}

			if resp.Token == "" {
				t.Error("Login() returned empty token")
			}

			// Verify token is valid
			_, err = service.ValidateToken(resp.Token)
			if err != nil {
				t.Errorf("Generated token is invalid: %v", err)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	repo := NewMockAuthRepository()
	jwtSecret := []byte("test-secret")
	service := NewAuthService(repo, jwtSecret)

	// Add a test user
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		CreatedAt: time.Now(),
	}
	repo.AddUser(user)

	// Generate a valid token
	validToken, err := service.generateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Generate an expired token
	expiredClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, err := expiredTokenObj.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to generate expired test token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		setup       func()
		wantErr     bool
		expectedID  int64
	}{
		{
			name:        "valid token",
			token:       validToken,
			setup:       func() {},
			wantErr:     false,
			expectedID:  1,
		},
		{
			name:    "empty token",
			token:   "",
			setup:   func() {},
			wantErr: true,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			setup:   func() {},
			wantErr: true,
		},
		{
			name:    "expired token",
			token:   expiredToken,
			setup:   func() {},
			wantErr: true,
		},
		{
			name:  "valid token but user not found",
			token: validToken,
			setup: func() {
				repo.SetGetUserError(errors.New("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset repository state
			repo.SetGetUserError(nil)
			tt.setup()

			user, err := service.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Error("ValidateToken() returned nil user on success")
					return
				}
				if user.ID != tt.expectedID {
					t.Errorf("ValidateToken() returned user ID = %v, want %v", user.ID, tt.expectedID)
				}
			}
		})
	}
}

func TestAuthService_generateToken(t *testing.T) {
	service := NewAuthService(NewMockAuthRepository(), []byte("test-secret"))

	user := &models.User{
		ID:       123,
		Email:    "test@example.com",
		Username: "testuser",
	}

	token, err := service.generateToken(user)
	if err != nil {
		t.Errorf("generateToken() error = %v", err)
		return
	}

	if token == "" {
		t.Error("generateToken() returned empty token")
		return
	}

	// Parse the token to verify its contents
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return service.jwtSecret, nil
	})

	if err != nil {
		t.Errorf("Failed to parse generated token: %v", err)
		return
	}

	if !parsedToken.Valid {
		t.Error("Generated token is not valid")
		return
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		t.Error("Failed to parse token claims")
		return
	}

	if claims.UserID != user.ID {
		t.Errorf("Token UserID = %v, want %v", claims.UserID, user.ID)
	}

	if claims.Email != user.Email {
		t.Errorf("Token Email = %v, want %v", claims.Email, user.Email)
	}

	// Verify expiration is set to 24 hours from now (with some tolerance)
	expectedExp := time.Now().Add(24 * time.Hour)
	actualExp := claims.ExpiresAt.Time
	diff := actualExp.Sub(expectedExp)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Token expiration is not approximately 24 hours from now. Expected: %v, Actual: %v", expectedExp, actualExp)
	}
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	service := NewAuthService(NewMockAuthRepository(), []byte("test-secret"))
	password := "StrongPass123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.hashPassword(password)
		if err != nil {
			b.Fatalf("hashPassword() failed: %v", err)
		}
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	service := NewAuthService(NewMockAuthRepository(), []byte("test-secret"))
	password := "StrongPass123!"
	hash, err := service.hashPassword(password)
	if err != nil {
		b.Fatalf("Failed to generate hash for benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.verifyPassword(hash, password)
	}
}