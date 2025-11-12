package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/manish-npx/go-echo-pg/internal/config"
	"github.com/manish-npx/go-echo-pg/internal/constants"
	"github.com/manish-npx/go-echo-pg/internal/model"
	"github.com/manish-npx/go-echo-pg/internal/repository"
	"github.com/manish-npx/go-echo-pg/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *model.CreateUserRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	GetUserProfile(ctx context.Context, userID pgtype.UUID) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userID pgtype.UUID, req *model.UpdateUserRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID pgtype.UUID, req *model.ChangePasswordRequest) error
}

type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
	logger   *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, config *config.Config, logger *zap.Logger) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   config,
		logger:   logger,
	}
}

func (s *authService) Register(ctx context.Context, req *model.CreateUserRequest) (*model.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New(constants.ErrUserExists)
	}

	// Create user
	user, err := s.userRepo.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Generate token
	token, expiresAt, err := utils.GenerateToken(user, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	s.logger.Info("User registered successfully",
		zap.String("email", user.Email),
		zap.String("user_id", user.ID.String()),
	)

	return &model.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("Login attempt with non-existent email", zap.String("email", req.Email))
		return nil, errors.New(constants.ErrInvalidCredentials)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.logger.Warn("Invalid password attempt", zap.String("email", req.Email))
		return nil, errors.New(constants.ErrInvalidCredentials)
	}

	token, expiresAt, err := utils.GenerateToken(user, s.config)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	s.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.String("user_id", user.ID.String()),
	)

	return &model.AuthResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *authService) GetUserProfile(ctx context.Context, userID pgtype.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user profile: %w", err)
	}

	return user, nil
}

func (s *authService) UpdateUserProfile(ctx context.Context, userID pgtype.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	// Cast to concrete type to access UpdateUser method
	repo, ok := s.userRepo.(*repository.UserRepositoryImpl)
	if !ok {
		return nil, errors.New("repository type assertion failed")
	}

	user, err := repo.UpdateUser(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("error updating user profile: %w", err)
	}

	s.logger.Info("User profile updated",
		zap.String("user_id", userID.String()),
		zap.String("email", user.Email),
	)

	return user, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID pgtype.UUID, req *model.ChangePasswordRequest) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return errors.New("invalid current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Update password using repository
	repo, ok := s.userRepo.(*repository.UserRepositoryImpl)
	if !ok {
		return errors.New("repository type assertion failed")
	}

	err = repo.UpdatePassword(ctx, userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	s.logger.Info("Password changed successfully", zap.String("user_id", userID.String()))
	return nil
}
