package controllers

import (
	"socialmedia/blacklist"
	"socialmedia/config"
	"socialmedia/models"
	"socialmedia/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Token   string        `json:"token,omitempty"` // Token is optional
	User    *UserResponse `json:"user,omitempty"`  // User is optional
}

type UserResponse struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
}

type RegisterInput struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with auto-generated username
// @Tags Auth
// @Accept json
// @Produce json
// @Param registerInput body RegisterInput true "Register Input"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} AuthResponse
// @Failure 500 {object} AuthResponse
// @Router /api/register [post]
func Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Check if a user with the email already exists.
	var existingUser models.User
	if err := models.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "User already exists",
		})
	}

	// Hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(AuthResponse{
			Status:  "error",
			Message: "Could not hash password",
		})
	}

	// Generate a username based on the provided name.
	username := utils.GenerateUsername(input.Name)

	// Create a new user. Bio and ProfilePicture will be empty by default.
	user := models.User{
		Email:    input.Email,
		Name:     input.Name,
		Username: username,
		Password: string(hashedPassword),
	}

	if err := models.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(AuthResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Generate a JWT token.
	token, err := generateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(AuthResponse{
			Status:  "error",
			Message: "Could not create token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		Status:  "success",
		Message: "User registered successfully",
		Token:   token,
	})
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login godoc
// @Summary Login user
// @Description Login a user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginInput body LoginInput true "Login Input"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} AuthResponse
// @Failure 401 {object} AuthResponse
// @Failure 500 {object} AuthResponse
// @Router /api/login [post]
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	// Look up the user by email.
	var user models.User
	if err := models.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid email or password",
		})
	}

	// Compare the provided password with the hashed password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid email or password",
		})
	}

	// Generate JWT token.
	token, err := generateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(AuthResponse{
			Status:  "error",
			Message: "Could not create token",
		})
	}

	userResponse := UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		Name:           user.Name,
		Username:       user.Username,
		ProfilePicture: user.ProfilePicture,
	}

	return c.JSON(AuthResponse{
		Status:  "success",
		Message: "Login successful",
		Token:   token,
		User:    &userResponse,
	})
}

func generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// GoogleLogin redirects the client to Google’s OAuth consent page.
func GoogleLogin(c *fiber.Ctx) error {
	url := utils.GetGoogleOAuthURL()
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

// GoogleCallback handles the callback from Google OAuth.
func GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Code not found"})
	}

	// Exchange the code for an access token and fetch user info.
	userInfo, err := utils.GetGoogleUserInfo(code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Check if a user with this email exists.
	var user models.User
	if err := models.DB.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		// If not, create a new user with auto‑generated username.
		username := utils.GenerateUsername(userInfo.Name)
		user = models.User{
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			Username:       username,
			ProfilePicture: userInfo.Picture,
		}
		models.DB.Create(&user)
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create token"})
	}

	return c.JSON(fiber.Map{"token": token, "user": user})
}

// Logout godoc
// @Summary Logout user
// @Description Logs out the user by adding their token to a blacklist so it cannot be used further.
// @Tags Auth
// @Produce json
// @Success 200 {object} AuthResponse
// @Failure 400 {object} AuthResponse
// @Router /api/logout [post]
func Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "Authorization header not found",
		})
	}

	// Expect token in format "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid authorization header",
		})
	}
	tokenStr := authHeader[len(bearerPrefix):]

	// Parse token to extract expiration (using the same secret).
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid token",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid token claims",
		})
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(AuthResponse{
			Status:  "error",
			Message: "Invalid expiration time",
		})
	}
	expirationTime := time.Unix(int64(expFloat), 0)

	// Add token to blacklist.
	blacklist.Add(tokenStr, expirationTime)

	return c.JSON(AuthResponse{
		Status:  "success",
		Message: "Logout successful",
	})
}
