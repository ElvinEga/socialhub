package controllers

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"time"

	"socialmedia/models"

	"github.com/gofiber/fiber/v2"
	openai "github.com/sashabaranov/go-openai"
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type Request struct {
	Content string `json:"content"`
}

type Prompt struct {
	Prompt string `json:"prompt"`
}

// CreateAIChatPost godoc
// @Summary Create an AI Chat Post
// @Description Create a new AI chat post with an initial prompt. The post type will be set to "ai".
// @Tags AIChat
// @Accept json
// @Produce json
// @Param request body Request true "Initial prompt from the user"
// @Success 201 {object} AIChatPostResponse
// @Failure 400 {object} ErrorResponse
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

	if body.Content == "" {
		// If the Content field is empty, return a 400 Bad Request status with an error message
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content cannot be empty"})
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
// @Param request body Request true "Chat message data"
// @Success 200 {object} ChatMessageResponse
// @Failure 400 {object} ErrorResponse
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
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
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

// SendAIChatMessage godoc
// @Summary Add new message to an AI chat post and get a response from OpenAI
// @Description For an existing AI chat post, this endpoint accepts a new user prompt, saves it, sends it to OpenAI, streams the response in real time (using Server-Sent Events), and finally appends the complete AI reply as a new message in the conversation thread.
// @Tags AIChat
// @Accept json
// @Produce text/event-stream
// @Param id path int true "AI Chat Post ID"
// @Param prompt body Prompt true "Prompt Message"
// @Success 200 "Streamed OpenAI response"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ai-posts/{id}/messages [post]
func SendAIChatMessage(c *fiber.Ctx) error {
	// Parse the AI Chat Post ID from the URL.
	postID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	// Retrieve the AI chat post.
	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}
	if post.PostType != "ai" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Post is not an AI chat post"})
	}

	// Parse the request body to get the user's prompt.
	type Request struct {
		Prompt string `json:"prompt"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Prompt cannot be empty"})
	}

	// Save the user's message as a ChatMessage.
	userMsg := models.ChatMessage{
		PostID:    post.ID,
		Sender:    "user",
		Content:   req.Prompt,
		CreatedAt: time.Now(),
	}
	if err := models.DB.Create(&userMsg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store user message: " + err.Error()})
	}

	// Set up the OpenAI client.
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "OPENAI_API_KEY is not set"})
	}
	client := openai.NewClient(openaiKey)

	// Build the conversation context.
	// (For a complete conversation, you might load previous messages. Here we use only the new prompt.)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Prompt,
		},
	}

	// Create a ChatCompletion request with streaming enabled.
	chatReq := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
		Stream:   true,
	}

	// Set header for Server-Sent Events.
	c.Set("Content-Type", "text/event-stream")

	// Variable to accumulate the full AI response.
	var fullResponse string

	// Stream the OpenAI response to the client.
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		stream, err := client.CreateChatCompletionStream(context.Background(), chatReq)
		if err != nil {
			w.WriteString("data: error: " + err.Error() + "\n\n")
			w.Flush()
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				w.WriteString("data: error: " + err.Error() + "\n\n")
				w.Flush()
				break
			}
			// Get the delta text from the response.
			chunk := response.Choices[0].Delta.Content
			if chunk != "" {
				fullResponse += chunk
				w.WriteString("data: " + chunk + "\n\n")
				w.Flush()
			}
		}
	})

	if err != nil {
		return err
	}

	// After streaming is complete, save the AI's complete reply as a ChatMessage.
	aiMsg := models.ChatMessage{
		PostID:    post.ID,
		Sender:    "ai",
		Content:   fullResponse,
		CreatedAt: time.Now(),
	}
	if err := models.DB.Create(&aiMsg).Error; err != nil {
		// Optionally log the error; the streaming response is already complete.
	}

	return nil
}
