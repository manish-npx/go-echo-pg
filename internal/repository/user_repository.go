package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/manish-npx/go-echo-pg/internal/constants"
	"github.com/manish-npx/go-echo-pg/internal/database"
	"github.com/manish-npx/go-echo-pg/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (*model.User, error)
}

// UserRepositoryImpl implements UserRepository
type UserRepositoryImpl struct {
	db     *database.DB
	logger *zap.Logger
}

func NewUserRepository(db *database.DB, logger *zap.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	query := `
		INSERT INTO users (email, password, name)
		VALUES ($1, $2, $3)
		RETURNING id, email, password, name, created_at, updated_at
	`

	var user model.User
	err = r.db.Pool.QueryRow(ctx, query, req.Email, string(hashedPassword), req.Name).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New(constants.ErrUserExists)
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	r.logger.Info("User created successfully", zap.String("email", user.Email))
	return &user, nil
}

func (r *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constants.ErrUserNotFound)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, id pgtype.UUID) (*model.User, error) {
	query := `
		SELECT id, email, password, name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constants.ErrUserNotFound)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

// Additional methods for update operations
func (r *UserRepositoryImpl) UpdateUser(ctx context.Context, id pgtype.UUID, req *model.UpdateUserRequest) (*model.User, error) {
	query := `
		UPDATE users
		SET name = $2, email = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, password, name, created_at, updated_at
	`

	var user model.User
	err := r.db.Pool.QueryRow(ctx, query, id, req.Name, req.Email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	r.logger.Info("User updated successfully", zap.String("email", user.Email))
	return &user, nil
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID pgtype.UUID, newPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Pool.Exec(ctx, query, newPassword, userID)
	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New(constants.ErrUserNotFound)
	}

	return nil
}
