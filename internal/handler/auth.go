package handler

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/manish-npx/go-echo-pg/internal/constants"
	"github.com/manish-npx/go-echo-pg/internal/model"
	"github.com/manish-npx/go-echo-pg/internal/service"
	"github.com/manish-npx/go-echo-pg/internal/utils"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
	response    *utils.ResponseHelper
	logger      *zap.Logger
}

func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		response:    utils.NewResponseHelper(logger),
		logger:      logger,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return h.response.BadRequest(c, "Invalid request format", err)
	}

	fmt.Println("req =================================>", req)
	if err := c.Validate(req); err != nil {
		return h.response.ValidationError(c, err.Error(), err)
	}

	authResponse, err := h.authService.Register(c.Request().Context(), &req)
	if err != nil {
		h.logger.Error("Registration failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)

		if err.Error() == constants.ErrUserExists {
			return h.response.Conflict(c, "User with this email already exists", err)
		}
		return h.response.BadRequest(c, "Registration failed", err)
	}

	return h.response.Created(c, authResponse)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return h.response.BadRequest(c, "Invalid request format", err)
	}

	if err := c.Validate(req); err != nil {
		return h.response.ValidationError(c, err.Error(), err)
	}

	authResponse, err := h.authService.Login(c.Request().Context(), &req)
	if err != nil {
		h.logger.Warn("Login failed",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		return h.response.Unauthorized(c, "Invalid email or password", err)
	}

	return h.response.Success(c, authResponse)
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID, ok := c.Get("userID").(pgtype.UUID)
	if !ok {
		return h.response.Unauthorized(c, "Invalid user ID", nil)
	}

	user, err := h.authService.GetUserProfile(c.Request().Context(), userID)
	if err != nil {
		return h.response.NotFound(c, "User not found", err)
	}

	return h.response.Success(c, user)
}

func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get("userID").(pgtype.UUID)
	if !ok {
		return h.response.Unauthorized(c, "Invalid user ID", nil)
	}

	var req model.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return h.response.BadRequest(c, "Invalid request format", err)
	}

	if err := c.Validate(req); err != nil {
		return h.response.ValidationError(c, err.Error(), err)
	}

	user, err := h.authService.UpdateUserProfile(c.Request().Context(), userID, &req)
	if err != nil {
		return h.response.BadRequest(c, "Failed to update profile", err)
	}

	return h.response.Success(c, user)
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	userID, ok := c.Get("userID").(pgtype.UUID)
	if !ok {
		return h.response.Unauthorized(c, "Invalid user ID", nil)
	}

	var req model.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return h.response.BadRequest(c, "Invalid request format", err)
	}

	if err := c.Validate(req); err != nil {
		return h.response.ValidationError(c, err.Error(), err)
	}

	err := h.authService.ChangePassword(c.Request().Context(), userID, &req)
	if err != nil {
		if err.Error() == "invalid current password" {
			return h.response.BadRequest(c, "Invalid current password", err)
		}
		return h.response.BadRequest(c, "Failed to change password", err)
	}

	return h.response.Success(c, map[string]string{"message": "Password changed successfully"})
}

func (h *AuthHandler) Health(c echo.Context) error {
	healthData := model.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Details: map[string]interface{}{
			"database": "connected",
			"redis":    "disabled",
		},
	}
	return h.response.Success(c, healthData)
}

func (h *AuthHandler) Ready(c echo.Context) error {
	healthData := model.HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	}
	return h.response.Success(c, healthData)
}
