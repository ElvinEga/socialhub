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

// RegisterInput represents the expected payload on registration.
type RegisterInput struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
	Bio            string `json:"bio"`
}

// Register a new user and auto‐generate a username.
func Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Check if a user with the email already exists.
	var existingUser models.User
	if err := models.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "User already exists"})
	}

	// Hash the password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not hash password"})
	}

	// Generate a username based on the provided name.
	username := utils.GenerateUsername(input.Name)

	user := models.User{
		Email:          input.Email,
		Password:       string(hashedPassword),
		Name:           input.Name,
		Username:       username,
		ProfilePicture: input.ProfilePicture,
		Bio:            input.Bio,
		CreatedAt:      time.Now(),
	}

	if err := models.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Generate a JWT token.
	token, err := generateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create token"})
	}

	return c.JSON(fiber.Map{"token": token, "user": user})
}

// LoginInput represents the expected payload for login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login validates credentials and returns a JWT token.
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var user models.User
	if err := models.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create token"})
	}

	return c.JSON(fiber.Map{"token": token, "user": user})
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
