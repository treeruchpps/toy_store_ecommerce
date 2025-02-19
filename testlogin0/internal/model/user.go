package model

import "time"

type User struct {
	ID                string    `json:"id" db:"user_id"`
	GoogleID          string    `json:"google_id" db:"google_id"`
	Email             string    `json:"email" db:"email"`
	FullName          string    `json:"full_name" db:"full_name"`
	DisplayName       string    `json:"display_name" db:"display_name"`
	Address           string    `json:"address" db:"address"`
	Phone             string    `json:"phone" db:"phone"`
	ProfilePictureURL string    `json:"profile_picture_url" db:"profile_picture_url"`
	EmailVerified     bool      `json:"email_verified" db:"email_verified"`
	Status            string    `json:"status" db:"status"`
	Role              string    `json:"role" db:"role"`
	LastLoginAt       time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        *User  `json:"user"`
}
