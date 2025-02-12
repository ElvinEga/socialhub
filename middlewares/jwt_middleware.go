package middlewares

import (
	"socialmedia/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func JWTMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or malformed JWT"})
	}

	// Expect token in format "Bearer <token>".
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization header format"})
	}
	tokenStr := authHeader[7:]

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Ensure token signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid JWT claims"})
	}

	// The user_id was set as a number (float64) during token generation.
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID in token"})
	}
	c.Locals("user_id", uint(userIDFloat))
	return c.Next()
}
