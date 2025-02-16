package controllers

import (
	"strconv"
	"time"

	"socialmedia/models"

	"github.com/gofiber/fiber/v2"
)

// AIChatPostResponse defines the response structure for an AI chat post.
type AIChatPostResponse struct {
	ID        uint                  `json:"id"`
	UserID    uint                  `json:"user_id"`
	PostType  string                `json:"post_type"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Messages  []ChatMessageResponse `json:"messages"`
}

// ChatMessageResponse defines the response structure for a chat message.
type ChatMessageResponse struct {
	ID        uint      `json:"id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAIChatPost godoc
// @Summary Create an AI Chat Post
// @Description Create a new AI chat post with an initial prompt. The post type will be set to "ai".
// @Tags AIChat
// @Accept json
// @Produce json
//
//	@Param request body struct {
//	    Content string `json:"content"`
//	} true "Initial prompt from the user"
//
// @Success 201 {object} AIChatPostResponse
// @Failure 400 {object} fiber.Map
// @Router /api/ai-posts [post]
func CreateAIChatPost(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	type Request struct {
		Content string `json:"content"`
	}
	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Create a new Post with PostType "ai"
	post := models.Post{
		Content:   "", // Not used for AI chat posts; conversation is in ChatMessage
		UserID:    userID,
		PostType:  "ai",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := models.DB.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Create the initial chat message (the user's prompt)
	message := models.ChatMessage{
		PostID:    post.ID,
		Sender:    "user",
		Content:   body.Content,
		CreatedAt: time.Now(),
	}
	if err := models.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := AIChatPostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		PostType:  post.PostType,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		Messages: []ChatMessageResponse{
			{
				ID:        message.ID,
				Sender:    message.Sender,
				Content:   message.Content,
				CreatedAt: message.CreatedAt,
			},
		},
	}
	return c.Status(fiber.StatusCreated).JSON(response)
}

// AddChatMessage godoc
// @Summary Add a Message to an AI Chat Post
// @Description Append a new chat message (either a prompt or reply) to an existing AI chat post.
// @Tags AIChat
// @Accept json
// @Produce json
// @Param id path int true "AI Chat Post ID"
//
//	@Param request body struct {
//	    Sender  string `json:"sender"`  // "user" or "ai"
//	    Content string `json:"content"`
//	} true "Chat message data"
//
// @Success 200 {object} ChatMessageResponse
// @Failure 400 {object} fiber.Map
// @Router /api/ai-posts/{id}/messages [post]
func AddChatMessage(c *fiber.Ctx) error {
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	type Request struct {
		Sender  string `json:"sender"`
		Content string `json:"content"`
	}
	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Verify the post exists and is an AI chat post.
	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}
	if post.PostType != "ai" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Post is not an AI chat post"})
	}

	message := models.ChatMessage{
		PostID:    uint(postID),
		Sender:    body.Sender,
		Content:   body.Content,
		CreatedAt: time.Now(),
	}
	if err := models.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	response := ChatMessageResponse{
		ID:        message.ID,
		Sender:    message.Sender,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
	return c.JSON(response)
}

// GetAIChatPost godoc
// @Summary Get an AI Chat Post
// @Description Retrieve an AI chat post along with its conversation thread.
// @Tags AIChat
// @Produce json
// @Param id path int true "AI Chat Post ID"
// @Success 200 {object} AIChatPostResponse
// @Failure 400 {object} fiber.Map
// @Failure 404 {object} fiber.Map
// @Router /api/ai-posts/{id} [get]
func GetAIChatPost(c *fiber.Ctx) error {
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}
	if post.PostType != "ai" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not an AI chat post"})
	}

	var messages []models.ChatMessage
	models.DB.Where("post_id = ?", post.ID).Order("created_at asc").Find(&messages)

	chatMessages := make([]ChatMessageResponse, len(messages))
	for i, m := range messages {
		chatMessages[i] = ChatMessageResponse{
			ID:        m.ID,
			Sender:    m.Sender,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		}
	}

	response := AIChatPostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		PostType:  post.PostType,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		Messages:  chatMessages,
	}
	return c.JSON(response)
}
