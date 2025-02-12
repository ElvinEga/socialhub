package controllers

import (
	"socialmedia/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// LikePost lets a user like a post.
func LikePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	// Check if the post has already been liked by this user.
	var like models.Like
	if err := models.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Already liked"})
	}

	like = models.Like{
		UserID: userID,
		PostID: uint(postID),
	}

	if err := models.DB.Create(&like).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(like)
}

// UnlikePost lets a user remove their like.
func UnlikePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var like models.Like
	if err := models.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Like not found"})
	}

	models.DB.Delete(&like)
	return c.JSON(fiber.Map{"message": "Unliked"})
}
