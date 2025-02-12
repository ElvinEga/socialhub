package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreatePost allows an authenticated user to create a new post.
func CreatePost(c *fiber.Ctx) error {
	// Retrieve the authenticated user’s ID (set by the JWT middleware)
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
func Timeline(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	// Load the current user with the “Following” relationship.
	var user models.User
	if err := models.DB.Preload("Following").First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Build a slice of user IDs: the current user plus all they follow.
	ids := []uint{userID}
	for _, u := range user.Following {
		ids = append(ids, u.ID)
	}

	var posts []models.Post
	models.DB.Preload("User").Where("user_id IN ?", ids).Order("created_at desc").Find(&posts)

	return c.JSON(posts)
}
