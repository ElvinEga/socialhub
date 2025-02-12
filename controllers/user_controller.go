package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetProfile returns the profile of the authenticated user.
func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user models.User
	if err := models.DB.Preload("Followers").Preload("Following").First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Build profile response.
	profile := fiber.Map{
		"id":              user.ID,
		"email":           user.Email,
		"name":            user.Name,
		"username":        user.Username,
		"profile_picture": user.ProfilePicture,
		"bio":             user.Bio,
		"created_at":      user.CreatedAt,
		"followers_count": len(user.Followers),
		"following_count": len(user.Following),
	}
	return c.JSON(profile)
}

// FollowUser lets the current user follow another user.
func FollowUser(c *fiber.Ctx) error {
	currentUserID := c.Locals("user_id").(uint)
	targetID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if uint(targetID) == currentUserID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot follow yourself"})
	}

	// Check if already following.
	var follow models.Follow
	if err := models.DB.Where("follower_id = ? AND following_id = ?", currentUserID, targetID).First(&follow).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Already following"})
	}

	follow = models.Follow{
		FollowerID:  currentUserID,
		FollowingID: uint(targetID),
		CreatedAt:   time.Now(),
	}

	if err := models.DB.Create(&follow).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User followed"})
}

// UnfollowUser lets the current user unfollow another user.
func UnfollowUser(c *fiber.Ctx) error {
	currentUserID := c.Locals("user_id").(uint)
	targetID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var follow models.Follow
	if err := models.DB.Where("follower_id = ? AND following_id = ?", currentUserID, targetID).First(&follow).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not following"})
	}

	if err := models.DB.Delete(&follow).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User unfollowed"})
}
