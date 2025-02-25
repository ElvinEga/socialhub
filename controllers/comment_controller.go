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

// AddReply godoc
// @Summary Add a reply to a comment
// @Description Allows a user to reply to an existing comment. The reply is also a comment, with a ParentID linking it to the original comment.
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Parent Comment ID"
//
//	@Param request body struct {
//	    Content string `json:"content"`
//	} true "Reply content"
//
// @Success 201 {object} models.Comment
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /comments/:id/replies [post]
func AddReply(c *fiber.Ctx) error {
	// Get the ID of the comment being replied to.
	parentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid comment ID"})
	}

	// Ensure the parent comment exists.
	var parentComment models.Comment
	if err := models.DB.First(&parentComment, parentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Parent comment not found"})
	}

	// Get the authenticated user ID.
	userID := c.Locals("user_id").(uint)

	// Parse the request body.
	type Request struct {
		Content string `json:"content"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content cannot be empty"})
	}

	// Create the reply comment.
	reply := models.Comment{
		Content:   req.Content,
		UserID:    userID,
		PostID:    parentComment.PostID, // The reply belongs to the same post.
		ParentID:  &parentComment.ID,    // Link to the parent comment.
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := models.DB.Create(&reply).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(reply)
}
