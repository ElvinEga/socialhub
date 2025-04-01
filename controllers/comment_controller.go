package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CommentResponse represents the response structure for comments
type CommentResponse struct {
	Comments []models.Comment   `json:"comments"`
	Metadata PaginationMetadata `json:"metadata"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int64 `json:"total_pages"`
}

// SingleCommentResponse represents the response for a single comment
type SingleCommentResponse struct {
	Comment       models.Comment  `json:"comment"`
	ParentComment *models.Comment `json:"parent_comment,omitempty"`
}

// AddComment godoc
// @Summary Add a comment to a post
// @Description Allows a user to comment on a post
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param request body Request true "Comment content"
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

	// Start a transaction
	tx := models.DB.Begin()

	comment := models.Comment{
		Content: input.Content,
		UserID:  userID,
		PostID:  uint(postID),
	}

	// Create the comment
	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: err.Error(),
		})
	}

	// Increment the comment count in the post table
	if err := tx.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).
		Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to update comment count",
		})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to commit transaction",
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
// @Param request body Request true "Comment content"
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

	// Start a transaction
	tx := models.DB.Begin()

	// Delete the comment
	if err := tx.Delete(&comment).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to delete comment",
		})
	}

	// Decrement the comment count in the post table
	if err := tx.Model(&models.Post{}).
		Where("id = ?", comment.PostID).
		UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", 1)).
		Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to update comment count",
		})
	}

	// If this was a parent comment, also delete all replies
	if comment.ParentCommentID == nil {
		// Delete all replies and update the comment count accordingly
		var replyCount int64
		if err := tx.Model(&models.Comment{}).
			Where("parent_id = ?", comment.ID).
			Count(&replyCount).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
				Message: "Failed to count replies",
			})
		}

		if replyCount > 0 {
			// Delete all replies
			if err := tx.Where("parent_id = ?", comment.ID).Delete(&models.Comment{}).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
					Message: "Failed to delete replies",
				})
			}

			// Decrease the comment count by the number of replies
			if err := tx.Model(&models.Post{}).
				Where("id = ?", comment.PostID).
				UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", replyCount)).
				Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
					Message: "Failed to update comment count for replies",
				})
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to commit transaction",
		})
	}

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
// @Param request body Request true "Reply content"
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
		Content:         req.Content,
		UserID:          userID,
		PostID:          parentComment.PostID, // The reply belongs to the same post.
		ParentCommentID: &parentComment.ID,    // Link to the parent comment.

	}
	if err := models.DB.Create(&reply).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(reply)
}

// GetCommentsByPostID godoc
// @Summary Get comments for a specific post
// @Description Get paginated comments for a post, including replies
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Comments per page (default: 10)"
// @Success 200 {object} CommentResponse
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /posts/{id}/comments [get]
func GetCommentsByPostID(c *fiber.Ctx) error {
	// Get post ID from URL parameters
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Invalid post ID",
		})
	}

	// Get pagination parameters
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Check if post exists
	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(MessageResponse{
			Message: "Post not found",
		})
	}

	var comments []models.Comment
	var total int64

	// Get total count of parent comments (comments without ParentID)
	if err := models.DB.Model(&models.Comment{}).
		Where("post_id = ? AND parent_id IS NULL", postID).
		Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to count comments",
		})
	}

	// Get paginated parent comments with their replies and user information
	if err := models.DB.
		Preload("User").         // Load comment author
		Preload("Replies").      // Load replies
		Preload("Replies.User"). // Load reply authors
		Where("post_id = ? AND parent_id IS NULL", postID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
			Message: "Failed to fetch comments",
		})
	}

	return c.JSON(fiber.Map{
		"comments": comments,
		"metadata": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCommentByID godoc
// @Summary Get a single comment by ID
// @Description Get a comment with its replies and user information
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} SingleCommentResponse
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /comments/{id} [get]
func GetCommentByID(c *fiber.Ctx) error {
	// Get comment ID from URL parameters
	commentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MessageResponse{
			Message: "Invalid comment ID",
		})
	}

	var comment models.Comment

	// Get the comment with all related data
	if err := models.DB.
		Preload("User").         // Load comment author
		Preload("Replies").      // Load replies
		Preload("Replies.User"). // Load reply authors
		First(&comment, commentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(MessageResponse{
			Message: "Comment not found",
		})
	}

	// If this is a reply (has ParentID), get the parent comment
	if comment.ParentCommentID != nil {
		var parentComment models.Comment
		if err := models.DB.
			Preload("User").
			First(&parentComment, comment.ParentCommentID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(MessageResponse{
				Message: "Failed to fetch parent comment",
			})
		}

		return c.JSON(fiber.Map{
			"comment":        comment,
			"parent_comment": parentComment,
		})
	}

	return c.JSON(fiber.Map{
		"comment": comment,
	})
}
