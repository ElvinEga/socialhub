package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AddComment allows a user to comment on a post.
func AddComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	comment := models.Comment{
		Content:   input.Content,
		UserID:    userID,
		PostID:    uint(postID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := models.DB.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(comment)
}

// EditComment allows the comment’s author to update it.
func EditComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	commentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid comment ID"})
	}

	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Comment not found"})
	}

	if comment.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized"})
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	comment.Content = input.Content
	comment.UpdatedAt = time.Now()

	models.DB.Save(&comment)
	return c.JSON(comment)
}

// DeleteComment allows the comment’s author to delete it.
func DeleteComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	commentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid comment ID"})
	}

	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Comment not found"})
	}

	if comment.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized"})
	}

	models.DB.Delete(&comment)
	return c.JSON(fiber.Map{"message": "Comment deleted"})
}
