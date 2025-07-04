package models

import (
	"time"
)

type User struct {
	ID        int64      `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	FirstName *string    `json:"first_name" db:"first_name"`
	LastName  *string    `json:"last_name" db:"last_name"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}
