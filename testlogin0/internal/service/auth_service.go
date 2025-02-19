// authen_service.go
package service

import (
	"context"
	"fmt"
	"log"

	"login/internal/config"
	"login/internal/model"
	"login/internal/repository"
	"login/pkg/utils"

	"google.golang.org/api/idtoken"
)

type AuthService struct {
	userRepo *repository.UserRepository
	Cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, Cfg: cfg}
}

func (s *AuthService) GetClientID() (string, error) {
	return s.Cfg.GoogleClientID, nil
}

func (s *AuthService) VerifyGoogleToken(ctx context.Context, idToken string) (*model.AuthResponse, error) {
	log.Println(idToken, s.Cfg.GoogleClientID)
	payload, err := idtoken.Validate(ctx, idToken, s.Cfg.GoogleClientID)
	if err != nil {
		log.Println("Google ID Token validation failed:", err)
		return nil, err
	}

	log.Println("Google ID Token validation successful:", payload)

	user, err := s.userRepo.GetUserByGoogleID(ctx, payload.Subject)
	if err != nil {
		log.Println("Error retrieving user by Google ID:", err)
		// return nil, err
	}

	if user == nil {
		user = &model.User{
			GoogleID:          payload.Subject,
			Email:             payload.Claims["email"].(string),
			FullName:          payload.Claims["name"].(string),
			ProfilePictureURL: payload.Claims["picture"].(string),
			EmailVerified:     payload.Claims["email_verified"].(bool),
			Status:            "active",
			Role:              "customer",
		}
		err = s.userRepo.CreateUser(ctx, user)
		if err != nil {
			log.Println("Error creating user:", err)
			return nil, err
		}
	}

	token, err := utils.GenerateToken(user.ID, s.Cfg.JWTSecret)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken: token,
		User:        user,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	// Implement any necessary logout logic
	return nil
}

// func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
// 	return s.userRepo.GetUserByID(ctx, userID)
// }

func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}
