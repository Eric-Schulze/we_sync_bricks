package auth

import (
	"errors"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
)

type AuthRepository struct {
	db models.DBService
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db models.DBService) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

// GetUserByEmail retrieves a user by email from database
func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	sql := `
		SELECT id, username, email, first_name, last_name, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`

	rows, err := r.db.Query(sql, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("user not found")
	}

	var user models.User
	err = rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID from database
func (r *AuthRepository) GetUserByID(id int64) (*models.User, error) {
	sql := `
		SELECT id, username, email, first_name, last_name, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	rows, err := r.db.Query(sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("user not found")
	}

	var user models.User
	err = rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user in the database
func (r *AuthRepository) CreateUser(req *RegisterRequest) (*models.User, error) {
	// Prepare nullable string pointers
	var firstName, lastName *string
	if req.FirstName != "" {
		firstName = &req.FirstName
	}
	if req.LastName != "" {
		lastName = &req.LastName
	}

	// SQL query to insert new user and return the created record
	sql := `
		INSERT INTO users (username, email, password, first_name, last_name, created_at) 
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP) 
		RETURNING id, username, email, first_name, last_name, created_at, updated_at
	`

	rows, err := r.db.Query(sql, req.Username, req.Email, req.Password, firstName, lastName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("failed to create user")
	}

	var user models.User
	err = rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmailOrUsername retrieves a user by email or username from database
func (r *AuthRepository) GetUserByEmailOrUsername(login string) (*models.User, error) {
	sql := `
		SELECT id, username, email, password, first_name, last_name, created_at, updated_at 
		FROM users 
		WHERE email = $1 OR username = $1
	`

	rows, err := r.db.Query(sql, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("user not found")
	}

	var user models.User
	err = rows.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUserLastLogin updates the user's last login timestamp
func (r *AuthRepository) UpdateUserLastLogin(userID int64) error {
	sql := `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = $1`

	_, err := r.db.ExecSQL(sql, userID)
	return err
}
