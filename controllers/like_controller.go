package controllers

import (
	"socialmedia/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type LikeResponse struct {
	Message    string `json:"message" example:"Post liked successfully"`
	LikesCount int64  `json:"likes_count" example:"42"`
}

// LikePost lets a user like a post.
// @Summary Like a post
// @Description Add a like to a post by the authenticated user
// @Tags likes
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} LikeResponse "Successfully liked the post"
// @Failure 400 {object} ErrorResponse "Invalid post ID or Already liked"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /posts/{id}/like [post]
func LikePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	// Start database transaction
	tx := models.DB.Begin()

	// Check if the post exists
	var post models.Post
	if err := tx.First(&post, postID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	// Check if the post has already been liked by this user
	var like models.Like
	if err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error; err == nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Already liked"})
	}

	// Create new like
	like = models.Like{
		UserID: userID,
		PostID: &post.ID,
	}

	// Save the like and increment the post's like count
	if err := tx.Create(&like).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Increment like count
	if err := tx.Model(&post).Update("likes_count", gorm.Expr("likes_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update like count"})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{
		"message":     "Post liked successfully",
		"likes_count": post.LikeCount + 1,
	})
}

// UnlikePost lets a user remove their like.
// @Summary Unlike a post
// @Description Remove a like from a post by the authenticated user
// @Tags likes
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} LikeResponse "Successfully unliked the post"
// @Failure 400 {object} ErrorResponse "Invalid post ID or Like not found"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /posts/{id}/like [delete]
func UnlikePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	// Start database transaction
	tx := models.DB.Begin()

	// Check if the post exists
	var post models.Post
	if err := tx.First(&post, postID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	// Find the like
	var like models.Like
	if err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Like not found"})
	}

	// Delete the like and decrement the post's like count
	if err := tx.Delete(&like).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete like"})
	}

	// Decrement like count, ensure it doesn't go below 0
	if err := tx.Model(&post).Update("likes_count", gorm.Expr("GREATEST(likes_count - ?, 0)", 1)).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update like count"})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{
		"message":     "Post unliked successfully",
		"likes_count": post.LikeCount - 1,
	})
}
