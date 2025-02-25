package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AddComment godoc
// @Summary Add a comment to a post
// @Description Allows a user to comment on a post
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
//
//	@Param request body struct {
//	    Content string `json:"content"`
//	} true "Comment content"
//
// @Success 201 {object} models.Comment
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /posts/{id}/comments [post]
// @Security ApiKeyAuth
func AddComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Invalid post ID",
		})
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: err.Error(),
		})
	}

	comment := models.Comment{
		Content:   input.Content,
		UserID:    userID,
		PostID:    uint(postID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := models.DB.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// EditComment godoc
// @Summary Edit a comment
// @Description Allows the comment’s author to update it
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
//
//	@Param request body struct {
//	    Content string `json:"content"`
//	} true "Updated comment content"
//
// @Success 200 {object} models.Comment
// @Failure 400 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /comments/{id} [put]
// @Security ApiKeyAuth
func EditComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	commentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Invalid comment ID",
		})
	}

	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(MessageResponse{
			Message: "Comment not found",
		})
	}

	if comment.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(MessageResponse{
			Message: "Not authorized",
		})
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: err.Error(),
		})
	}

	comment.Content = input.Content
	comment.UpdatedAt = time.Now()

	models.DB.Save(&comment)
	return c.JSON(comment)
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Allows the comment’s author to delete it
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /comments/{id} [delete]
// @Security ApiKeyAuth
func DeleteComment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	commentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Invalid comment ID",
		})
	}

	var comment models.Comment
	if err := models.DB.First(&comment, commentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(MessageResponse{
			Message: "Comment not found",
		})
	}

	if comment.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(MessageResponse{
			Message: "Not authorized",
		})
	}

	models.DB.Delete(&comment)
	return c.JSON(MessageResponse{
		Message: "Comment deleted",
	})
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
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /comments/{id}/replies [post]
// @Security ApiKeyAuth
func AddReply(c *fiber.Ctx) error {
	// Get the ID of the comment being replied to.
	parentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Invalid comment ID",
		})
	}

	// Ensure the parent comment exists.
	var parentComment models.Comment
	if err := models.DB.First(&parentComment, parentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(MessageResponse{
			Message: "Parent comment not found",
		})
	}

	// Get the authenticated user ID.
	userID := c.Locals("user_id").(uint)

	// Parse the request body.
	type Request struct {
		Content string `json:"content"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: err.Error(),
		})
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Content cannot be empty",
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(reply)
}
