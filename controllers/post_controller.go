package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PostInput struct {
	Content  string `json:"content"`
	ImageURL string `json:"image_url"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// CreatePost allows an authenticated user to create a new post.
// @Summary Create a new post
// @Description Create a new post with content and optional image URL
// @Tags posts
// @Accept json
// @Produce json
// @Param postInput body PostInput true "Post Input"
// @Success 200 {object} models.Post
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [post]
func CreatePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	type PostInput struct {
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	}

	var input PostInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	post := models.Post{
		Content:   input.Content,
		ImageURL:  input.ImageURL,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := models.DB.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(post)
}

// EditPost allows the owner to update a post.
// @Summary Edit a post
// @Description Edit a post's content and image URL
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param postInput body PostInput true "Post Input"
// @Success 200 {object} models.Post
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /posts/{id} [put]
func EditPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	if post.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized"})
	}

	var input struct {
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	post.Content = input.Content
	post.ImageURL = input.ImageURL
	post.UpdatedAt = time.Now()

	models.DB.Save(&post)
	return c.JSON(post)
}

// DeletePost allows the owner to delete a post.
// @Summary Delete a post
// @Description Delete a post by ID
// @Tags posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /posts/{id} [delete]
func DeletePost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	if post.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized"})
	}

	models.DB.Delete(&post)
	return c.JSON(fiber.Map{"message": "Post deleted"})
}

// Timeline returns posts created by the authenticated user and those they follow.
// @Summary Get timeline posts
// @Description Get posts created by the authenticated user and those they follow
// @Tags posts
// @Produce json
// @Success 200 {array} models.Post
// @Failure 404 {object} ErrorResponse
// @Router /posts/timeline [get]
func Timeline(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user models.User
	if err := models.DB.Preload("Following").First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	ids := []uint{userID}
	for _, u := range user.Following {
		ids = append(ids, u.ID)
	}

	var posts []models.Post
	models.DB.Preload("User").Where("user_id IN ?", ids).Order("created_at desc").Find(&posts)

	return c.JSON(posts)
}

// PostList returns all posts with optional pagination.
// @Summary List all posts
// @Description Get a list of all posts with optional pagination
// @Tags posts
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Posts per page (default: 10)"
// @Success 200 {array} models.Post
// @Failure 400 {object} ErrorResponse
// @Router /posts [get]
func PostList(c *fiber.Ctx) error {
	// Get page and limit from query parameters
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	var posts []models.Post
	var total int64

	// Get total count of posts
	if err := models.DB.Model(&models.Post{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count posts",
		})
	}

	// Get posts with pagination
	if err := models.DB.
		Preload("User"). // Include user information
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch posts",
		})
	}

	// Return response with pagination metadata
	return c.JSON(fiber.Map{
		"posts": posts,
		"metadata": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
