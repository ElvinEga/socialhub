package controllers

import (
	"socialmedia/config"
	"socialmedia/models"
	"socialmedia/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"` // Token is optional
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
		Email:     input.Email,
		Name:      input.Name,
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		// Bio and ProfilePicture are not set at registration.
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

	return c.JSON(AuthResponse{
		Status:  "success",
		Message: "Login successful",
		Token:   token,
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
			CreatedAt:      time.Now(),
		}
		models.DB.Create(&user)
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create token"})
	}

	return c.JSON(fiber.Map{"token": token, "user": user})
}
