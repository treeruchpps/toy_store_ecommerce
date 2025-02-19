package repository

import (
	"context"
	"database/sql"

	"login/internal/model"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (google_id, email, full_name, display_name, address, phone, profile_picture_url, email_verified, status, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING user_id
	`
	return r.db.QueryRowContext(ctx, query,
		user.GoogleID, user.Email, user.FullName, user.FullName, user.Address, user.Phone, user.ProfilePictureURL,
		user.EmailVerified, user.Status, user.Role,
	).Scan(&user.ID)
}

func (r *UserRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	var user model.User
	query := "SELECT * FROM users WHERE google_id = $1"
	err := r.db.GetContext(ctx, &user, query, googleID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $2, full_name = $3, display_name = $4, address = $5, phone = $6, profile_picture_url = $7, email_verified = $8,
			status = $9, role = $10, last_login_at = $11, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.FullName, user.DisplayName, user.Address, user.Phone, user.ProfilePictureURL,
		user.EmailVerified, user.Status, user.Role, user.LastLoginAt,
	)
	return err
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	query := "SELECT * FROM users WHERE user_id = $1"
	err := r.db.GetContext(ctx, &user, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}
