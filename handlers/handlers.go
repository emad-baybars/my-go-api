package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go-backend-template/config"
	"go-backend-template/database"
	"go-backend-template/jwt"
	"go-backend-template/models"
	"go-backend-template/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	mongoDB       *database.MongoDB
	postgresDB    *database.PostgresDB
	logger        utils.Logger
	localizer     *utils.Localizer
	passwordUtils *utils.PasswordUtils
	jwtUtils      *utils.JWTUtils
	responseUtils *utils.ResponseUtils
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(mongoDB *database.MongoDB, postgresDB *database.PostgresDB, logger utils.Logger, localizer *utils.Localizer) *AuthHandler {
	cfg := config.Load()
	return &AuthHandler{
		mongoDB:       mongoDB,
		postgresDB:    postgresDB,
		logger:        logger,
		localizer:     localizer,
		passwordUtils: &utils.PasswordUtils{},
		jwtUtils:      utils.NewJWTUtils(cfg.JWTSecret),
		responseUtils: &utils.ResponseUtils{},
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, username, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration data"
// @Success 201 {object} models.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	lang := c.GetString("language")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Registration validation failed", "error", err)
		c.JSON(http.StatusBadRequest, h.responseUtils.ErrorResponse(
			h.localizer.Get(lang, "validation_error"),
			err.Error(),
		))
		return
	}

	// Hash password
	hashedPassword, err := h.passwordUtils.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Password hashing failed", "error", err)
		c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
			h.localizer.Get(lang, "internal_error"),
			"Failed to process password",
		))
		return
	}

	// Check if using PostgreSQL
	if h.postgresDB != nil {
		user := models.User{
			Email:     req.Email,
			Username:  req.Username,
			Password:  hashedPassword,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Role:      "user",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Check if user exists
		var existingUser models.User
		if err := h.postgresDB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "email_exists"),
				"User already exists",
			))
			return
		}

		// Create user
		if err := h.postgresDB.Create(&user).Error; err != nil {
			h.logger.Error("Failed to create user in PostgreSQL", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to create user",
			))
			return
		}

		// Generate token
		token, expiresAt, err := jwt.GenerateToken(h.jwtUtils.Secret, user.ID, user.Email, user.Username, user.Role)
		if err != nil {
			h.logger.Error("Token generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to generate token",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		authResponse := models.AuthResponse{
			Token:     token,
			User:      userInfo,
			ExpiresAt: expiresAt,
		}

		c.JSON(http.StatusCreated, h.responseUtils.SuccessResponse(
			h.localizer.Get(lang, "user_created"),
			authResponse,
		))
		return
	}

	// MongoDB implementation
	if h.mongoDB != nil {
		userMongo := models.UserMongo{
			Email:     req.Email,
			Username:  req.Username,
			Password:  hashedPassword,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Role:      "user",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Check if user exists
		collection := h.mongoDB.Collection("users")
		filter := bson.M{
			"$or": []bson.M{
				{"email": req.Email},
				{"username": req.Username},
			},
		}

		var existingUser models.UserMongo
		if err := collection.FindOne(context.Background(), filter).Decode(&existingUser); err == nil {
			c.JSON(http.StatusConflict, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "email_exists"),
				"User already exists",
			))
			return
		}

		// Create user
		result, err := collection.InsertOne(context.Background(), userMongo)
		if err != nil {
			h.logger.Error("Failed to create user in MongoDB", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to create user",
			))
			return
		}

		userMongo.ID = result.InsertedID.(primitive.ObjectID)

		// Generate token
		token, expiresAt, err := jwt.GenerateToken(h.jwtUtils.Secret, userMongo.ID.Hex(), userMongo.Email, userMongo.Username, userMongo.Role)
		if err != nil {
			h.logger.Error("Token generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to generate token",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        userMongo.ID.Hex(),
			Email:     userMongo.Email,
			Username:  userMongo.Username,
			FirstName: userMongo.FirstName,
			LastName:  userMongo.LastName,
			Role:      userMongo.Role,
			IsActive:  userMongo.IsActive,
			CreatedAt: userMongo.CreatedAt,
			UpdatedAt: userMongo.UpdatedAt,
		}

		authResponse := models.AuthResponse{
			Token:     token,
			User:      userInfo,
			ExpiresAt: expiresAt,
		}

		c.JSON(http.StatusCreated, h.responseUtils.SuccessResponse(
			h.localizer.Get(lang, "user_created"),
			authResponse,
		))
	}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.APIResponse{data=models.AuthResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	lang := c.GetString("language")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Login validation failed", "error", err)
		c.JSON(http.StatusBadRequest, h.responseUtils.ErrorResponse(
			h.localizer.Get(lang, "validation_error"),
			err.Error(),
		))
		return
	}

	// PostgreSQL implementation
	if h.postgresDB != nil {
		var user models.User
		if err := h.postgresDB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			h.logger.Error("User not found in PostgreSQL", "email", req.Email)
			c.JSON(http.StatusUnauthorized, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "invalid_credentials"),
				"Authentication failed",
			))
			return
		}

		// Verify password
		if err := h.passwordUtils.VerifyPassword(user.Password, req.Password); err != nil {
			h.logger.Error("Password verification failed", "email", req.Email)
			c.JSON(http.StatusUnauthorized, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "invalid_credentials"),
				"Authentication failed",
			))
			return
		}

		// Generate token
		token, expiresAt, err := jwt.GenerateToken(h.jwtUtils.Secret, user.ID, user.Email, user.Username, user.Role)
		if err != nil {
			h.logger.Error("Token generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to generate token",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		authResponse := models.AuthResponse{
			Token:     token,
			User:      userInfo,
			ExpiresAt: expiresAt,
		}

		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse(
			h.localizer.Get(lang, "login_successful"),
			authResponse,
		))
		return
	}

	// MongoDB implementation
	if h.mongoDB != nil {
		collection := h.mongoDB.Collection("users")
		filter := bson.M{"email": req.Email}

		var user models.UserMongo
		if err := collection.FindOne(context.Background(), filter).Decode(&user); err != nil {
			h.logger.Error("User not found in MongoDB", "email", req.Email)
			c.JSON(http.StatusUnauthorized, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "invalid_credentials"),
				"Authentication failed",
			))
			return
		}

		// Verify password
		if err := h.passwordUtils.VerifyPassword(user.Password, req.Password); err != nil {
			h.logger.Error("Password verification failed", "email", req.Email)
			c.JSON(http.StatusUnauthorized, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "invalid_credentials"),
				"Authentication failed",
			))
			return
		}

		// Generate token
		token, expiresAt, err := jwt.GenerateToken(h.jwtUtils.Secret, user.ID.Hex(), user.Email, user.Username, user.Role)
		if err != nil {
			h.logger.Error("Token generation failed", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to generate token",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID.Hex(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		authResponse := models.AuthResponse{
			Token:     token,
			User:      userInfo,
			ExpiresAt: expiresAt,
		}

		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse(
			h.localizer.Get(lang, "login_successful"),
			authResponse,
		))
	}
}

// UserHandler handles user-related requests
type UserHandler struct {
	mongoDB       *database.MongoDB
	postgresDB    *database.PostgresDB
	logger        utils.Logger
	localizer     *utils.Localizer
	responseUtils *utils.ResponseUtils
}

// NewUserHandler creates a new user handler
func NewUserHandler(mongoDB *database.MongoDB, postgresDB *database.PostgresDB, logger utils.Logger, localizer *utils.Localizer) *UserHandler {
	return &UserHandler{
		mongoDB:       mongoDB,
		postgresDB:    postgresDB,
		logger:        logger,
		localizer:     localizer,
		responseUtils: &utils.ResponseUtils{},
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the current user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} models.APIResponse{data=models.UserInfo}
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	lang := c.GetString("language")

	// PostgreSQL implementation
	if h.postgresDB != nil {
		var user models.User
		id, _ := strconv.ParseUint(userID, 10, 32)
		if err := h.postgresDB.First(&user, uint(id)).Error; err != nil {
			h.logger.Error("User not found in PostgreSQL", "user_id", userID)
			c.JSON(http.StatusNotFound, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "user_not_found"),
				"User not found",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse("Profile retrieved successfully", userInfo))
		return
	}

	// MongoDB implementation
	if h.mongoDB != nil {
		collection := h.mongoDB.Collection("users")
		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			h.logger.Error("Invalid user ID format", "user_id", userID)
			c.JSON(http.StatusBadRequest, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "bad_request"),
				"Invalid user ID format",
			))
			return
		}

		var user models.UserMongo
		if err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user); err != nil {
			h.logger.Error("User not found in MongoDB", "user_id", userID)
			c.JSON(http.StatusNotFound, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "user_not_found"),
				"User not found",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID.Hex(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse("Profile retrieved successfully", userInfo))
	}
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the current user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.UpdateUserRequest true "User update data"
// @Success 200 {object} models.APIResponse{data=models.UserInfo}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req models.UpdateUserRequest
	userID := c.GetString("user_id")
	lang := c.GetString("language")

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Profile update validation failed", "error", err)
		c.JSON(http.StatusBadRequest, h.responseUtils.ErrorResponse(
			h.localizer.Get(lang, "validation_error"),
			err.Error(),
		))
		return
	}

	// PostgreSQL implementation
	if h.postgresDB != nil {
		var user models.User
		id, _ := strconv.ParseUint(userID, 10, 32)
		if err := h.postgresDB.First(&user, uint(id)).Error; err != nil {
			h.logger.Error("User not found in PostgreSQL", "user_id", userID)
			c.JSON(http.StatusNotFound, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "user_not_found"),
				"User not found",
			))
			return
		}

		// Update fields
		if req.FirstName != "" {
			user.FirstName = req.FirstName
		}
		if req.LastName != "" {
			user.LastName = req.LastName
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		user.UpdatedAt = time.Now()

		if err := h.postgresDB.Save(&user).Error; err != nil {
			h.logger.Error("Failed to update user in PostgreSQL", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to update profile",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse(
			h.localizer.Get(lang, "user_updated"),
			userInfo,
		))
		return
	}

	// MongoDB implementation
	if h.mongoDB != nil {
		collection := h.mongoDB.Collection("users")
		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			h.logger.Error("Invalid user ID format", "user_id", userID)
			c.JSON(http.StatusBadRequest, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "bad_request"),
				"Invalid user ID format",
			))
			return
		}

		update := bson.M{
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		}

		if req.FirstName != "" {
			update["$set"].(bson.M)["first_name"] = req.FirstName
		}
		if req.LastName != "" {
			update["$set"].(bson.M)["last_name"] = req.LastName
		}
		if req.Email != "" {
			update["$set"].(bson.M)["email"] = req.Email
		}

		_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
		if err != nil {
			h.logger.Error("Failed to update user in MongoDB", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to update profile",
			))
			return
		}

		// Get updated user
		var user models.UserMongo
		if err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user); err != nil {
			h.logger.Error("Failed to retrieve updated user", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to retrieve updated profile",
			))
			return
		}

		userInfo := models.UserInfo{
			ID:        user.ID.Hex(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse(
			h.localizer.Get(lang, "user_updated"),
			userInfo,
		))
	}
}

// GetUsers godoc
// @Summary Get all users (Admin only)
// @Description Get paginated list of all users
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort query string false "Sort order" default("created_at:desc")
// @Param search query string false "Search term"
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse}
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	var query models.PaginationQuery
	lang := c.GetString("language")

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, h.responseUtils.ErrorResponse(
			h.localizer.Get(lang, "validation_error"),
			err.Error(),
		))
		return
	}

	// PostgreSQL implementation
	if h.postgresDB != nil {
		var users []models.User
		var total int64

		db := h.postgresDB.Model(&models.User{})

		// Apply search filter
		if query.Search != "" {
			searchPattern := "%" + query.Search + "%"
			db = db.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR username ILIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern)
		}

		// Count total records
		db.Count(&total)

		// Apply pagination
		offset := (query.Page - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)

		// Apply sorting
		if query.Sort != "" {
			db = db.Order(query.Sort)
		} else {
			db = db.Order("created_at DESC")
		}

		if err := db.Find(&users).Error; err != nil {
			h.logger.Error("Failed to retrieve users from PostgreSQL", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to retrieve users",
			))
			return
		}

		// Convert to UserInfo
		userInfos := make([]models.UserInfo, len(users))
		for i, user := range users {
			userInfos[i] = models.UserInfo{
				ID:        user.ID,
				Email:     user.Email,
				Username:  user.Username,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Role:      user.Role,
				IsActive:  user.IsActive,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
		}

		pagination := models.Pagination{
			Page:      query.Page,
			PageSize:  query.PageSize,
			Total:     total,
			TotalPage: int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
		}

		response := h.responseUtils.PaginatedResponse(userInfos, pagination)
		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse("Users retrieved successfully", response))
		return
	}

	// MongoDB implementation
	if h.mongoDB != nil {
		collection := h.mongoDB.Collection("users")
		ctx := context.Background()

		// Build filter
		filter := bson.M{}
		if query.Search != "" {
			filter = bson.M{
				"$or": []bson.M{
					{"first_name": bson.M{"$regex": query.Search, "$options": "i"}},
					{"last_name": bson.M{"$regex": query.Search, "$options": "i"}},
					{"email": bson.M{"$regex": query.Search, "$options": "i"}},
					{"username": bson.M{"$regex": query.Search, "$options": "i"}},
				},
			}
		}

		// Count total documents
		total, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			h.logger.Error("Failed to count users in MongoDB", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to count users",
			))
			return
		}

		// Find documents with pagination
		skip := int64((query.Page - 1) * query.PageSize)
		limit := int64(query.PageSize)

		cursor, err := collection.Find(ctx, filter, options.Find().SetSkip(skip).SetLimit(limit))
		if err != nil {
			h.logger.Error("Failed to retrieve users from MongoDB", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to retrieve users",
			))
			return
		}
		defer cursor.Close(ctx)

		var users []models.UserMongo
		if err := cursor.All(ctx, &users); err != nil {
			h.logger.Error("Failed to decode users from MongoDB", "error", err)
			c.JSON(http.StatusInternalServerError, h.responseUtils.ErrorResponse(
				h.localizer.Get(lang, "internal_error"),
				"Failed to decode users",
			))
			return
		}

		// Convert to UserInfo
		userInfos := make([]models.UserInfo, len(users))
		for i, user := range users {
			userInfos[i] = models.UserInfo{
				ID:        user.ID.Hex(),
				Email:     user.Email,
				Username:  user.Username,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Role:      user.Role,
				IsActive:  user.IsActive,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
		}

		pagination := models.Pagination{
			Page:      query.Page,
			PageSize:  query.PageSize,
			Total:     total,
			TotalPage: int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
		}

		response := h.responseUtils.PaginatedResponse(userInfos, pagination)
		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse("Users retrieved successfully", response))
	}
}

// HealthHandler handles health check requests
type HealthHandler struct {
	mongoDB       *database.MongoDB
	postgresDB    *database.PostgresDB
	logger        utils.Logger
	responseUtils *utils.ResponseUtils
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(mongoDB *database.MongoDB, postgresDB *database.PostgresDB, logger utils.Logger) *HealthHandler {
	return &HealthHandler{
		mongoDB:       mongoDB,
		postgresDB:    postgresDB,
		logger:        logger,
		responseUtils: &utils.ResponseUtils{},
	}
}

// HealthCheck godoc
// @Summary Health check
// @Description Check the health status of the API and connected services
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.HealthResponse}
// @Failure 500 {object} models.APIResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	services := make(map[string]string)
	overallStatus := "healthy"

	// Check PostgreSQL
	if h.postgresDB != nil {
		if err := h.postgresDB.HealthCheck(); err != nil {
			services["postgresql"] = "unhealthy"
			overallStatus = "unhealthy"
			h.logger.Error("PostgreSQL health check failed", "error", err)
		} else {
			services["postgresql"] = "healthy"
		}
	}

	// Check MongoDB
	if h.mongoDB != nil {
		if err := h.mongoDB.HealthCheck(); err != nil {
			services["mongodb"] = "unhealthy"
			overallStatus = "unhealthy"
			h.logger.Error("MongoDB health check failed", "error", err)
		} else {
			services["mongodb"] = "healthy"
		}
	}

	healthResponse := models.HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
		Version:   "1.0.0",
	}

	if overallStatus == "healthy" {
		c.JSON(http.StatusOK, h.responseUtils.SuccessResponse("System is healthy", healthResponse))
	} else {
		c.JSON(http.StatusServiceUnavailable, h.responseUtils.ErrorResponse("System is unhealthy", "One or more services are down"))
	}
}
