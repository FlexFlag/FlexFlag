package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/flexflag/flexflag/internal/auth"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo *postgres.UserRepository
}

func NewUserHandler(userRepo *postgres.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GeneratePassword generates a random password
func GeneratePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to base64 and trim to desired length
	password := base64.URLEncoding.EncodeToString(bytes)[:length]
	return password, nil
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with generated or provided password
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body types.CreateUserRequest true "User creation request"
// @Success 201 {object} map[string]interface{} "User created successfully with password"
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req types.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	exists, err := h.userRepo.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email existence"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Generate password if not provided
	plainPassword := req.Password
	if plainPassword == "" {
		plainPassword, err = GeneratePassword(12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate password"})
			return
		}
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default role if not specified
	if req.Role == "" {
		req.Role = types.UserRoleViewer
	}

	// Create user
	user := &types.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, gin.H{
		"user":     user,
		"password": plainPassword, // Return the plain password only on creation
		"message":  "User created successfully. Please save the password securely.",
	})
}

// ListUsers godoc
// @Summary List all users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}

	users, err := h.userRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	count, err := h.userRepo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": count,
		"limit": limit,
		"offset": offset,
	})
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get user details by ID
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} types.User
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	var user types.User
	if err := h.userRepo.GetByID(c.Request.Context(), id, &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user details (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param user body types.User true "User update request"
// @Success 200 {object} types.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var updateReq struct {
		FullName string         `json:"full_name"`
		Role     types.UserRole `json:"role"`
		IsActive bool          `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user types.User
	if err := h.userRepo.GetByID(c.Request.Context(), id, &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user fields
	if updateReq.FullName != "" {
		user.FullName = updateReq.FullName
	}
	if updateReq.Role != "" {
		user.Role = updateReq.Role
	}
	user.IsActive = updateReq.IsActive

	if err := h.userRepo.Update(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Soft delete a user (sets is_active to false)
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.userRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// UpdateProfile godoc
// @Summary Update own profile
// @Description Update current user's profile (full name only)
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param profile body map[string]string true "Profile update request"
// @Success 200 {object} types.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var updateReq struct {
		FullName string `json:"full_name"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateReq.FullName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Full name is required"})
		return
	}

	var user types.User
	if err := h.userRepo.GetByID(c.Request.Context(), userID.(string), &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update only full name
	user.FullName = updateReq.FullName

	if err := h.userRepo.Update(c.Request.Context(), &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// ResetPassword godoc
// @Summary Reset user password
// @Description Generate a new password for a user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "New password"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")

	// Check if user exists
	var user types.User
	if err := h.userRepo.GetByID(c.Request.Context(), id, &user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate new password
	newPassword, err := GeneratePassword(12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate password"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := h.userRepo.UpdatePassword(c.Request.Context(), id, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Password reset successfully",
		"password": newPassword,
	})
}
