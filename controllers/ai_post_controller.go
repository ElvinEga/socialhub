package controllers

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"socialmedia/models"

	openai "github.com/ElvinEga/go-openai"
	"github.com/gofiber/fiber/v2"
)

// AIChatPostResponse defines the response structure for an AI chat post.
type AIChatPostResponse struct {
	ID        uint                  `json:"id"`
	UserID    uint                  `json:"user_id"`
	PostType  string                `json:"post_type"`
	Content   string                `json:"content"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Messages  []ChatMessageResponse `json:"messages"`
	Media     []models.Media        `json:"media"`
}

// ChatMessageResponse defines the response structure for a chat message.
type ChatMessageResponse struct {
	ID        uint      `json:"id"`
	PostID    uint      `json:"post_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Request struct {
	Content string `json:"content"`
	Files   []*multipart.FileHeader `form:"files"`
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

    // Parse multipart form
    if err := c.ParseMultipartForm(); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Failed to parse multipart form",
        })
    }

    content := c.FormValue("content", "")
    if content == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Content cannot be empty",
        })
    }

    // Start a transaction
    tx := models.DB.Begin()

    // Create a new Post with PostType "ai"
    post := models.Post{
        Content:  content,
        UserID:   userID,
        PostType: "ai",
    }

    if err := tx.Create(&post).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    // Handle file uploads if any
    if form, err := c.MultipartForm(); err == nil && form.File != nil {
        files := form.File["files"]
        if len(files) > 0 {
            uploadcareService := services.GetUploadcareService()

            for _, file := range files {
                // Upload file to Uploadcare
                uploadResult, err := uploadcareService.UploadFile(file)
                if err != nil {
                    tx.Rollback()
                    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                        "error": fmt.Sprintf("Failed to upload file: %v", err),
                    })
                }

                // Create Media record
                mediaType := services.DetermineMediaType(uploadResult.MimeType)
                mediaItem := models.Media{
                    PostID:  post.ID,
                    UserID:  userID,
                    URL:     uploadResult.URL,
                    Type:    mediaType,
                    AltText: filepath.Base(file.Filename),
                }

                if err := tx.Create(&mediaItem).Error; err != nil {
                    tx.Rollback()
                    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                        "error": "Failed to save media item",
                    })
                }
            }
        }
    }

    // Commit the transaction
    if err := tx.Commit().Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to commit transaction",
        })
    }

    // Reload the post with all relationships
    var completePost models.Post
    if err := models.DB.Preload("Media").First(&completePost, post.ID).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to load post relationships",
        })
    }

    response := AIChatPostResponse{
        ID:        completePost.ID,
        UserID:    completePost.UserID,
        PostType:  completePost.PostType,
        Content:   completePost.Content,
        CreatedAt: completePost.CreatedAt,
        UpdatedAt: completePost.UpdatedAt,
        Messages:  []ChatMessageResponse{},
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
		PostID:  uint(postID),
		Sender:  body.Sender,
		Content: body.Content,
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
    if err := models.DB.Preload("Media").First(&post, postID).Error; err != nil {
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
			PostID:    m.PostID,
			Sender:    m.Sender,
			Content:   m.Content,
			Reason:    m.Reason,
			CreatedAt: m.CreatedAt,
		}
	}

	response := AIChatPostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		PostType:  post.PostType,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		Messages:  chatMessages,
		Media:     post.Media,
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
// @Param request body Request true "Chat message data"
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
		Content string `json:"content"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Prompt cannot be empty"})
	}

	var messageCount int64
	if err := models.DB.Model(&models.ChatMessage{}).Where("post_id = ?", post.ID).Count(&messageCount).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count messages: " + err.Error()})
	}

	// Save the user's message as a ChatMessage.
	userMsg := models.ChatMessage{
		PostID:  post.ID,
		Sender:  "user",
		Content: req.Content,
	}
	if err := models.DB.Create(&userMsg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store user message: " + err.Error()})
	}

	// Set up the OpenAI client.
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openrouterKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "OPENROUTER_API_KEY is not set"})
	}

	config := openai.DefaultConfig(openrouterKey)
	config.BaseURL = "https://openrouter.ai/api/v1" // Override the base URL for OpenRouter
	client := openai.NewClientWithConfig(config)

	// Build the conversation context.
	// (For a complete conversation, you might load previous messages. Here we use only the new prompt.)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Content,
		},
	}

	// Create a ChatCompletion request with streaming enabled.
	chatReq := openai.ChatCompletionRequest{
		Model:    "qwen/qwq-32b:free", // Use the desired OpenRouter model
		Messages: messages,
		Stream:   true,
	}

	// Set header for Server-Sent Events.
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// Variable to accumulate the full AI response.
	var fullResponse string
	var reasonResponse string

	// Stream the OpenAI response to the client.
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		stream, err := client.CreateChatCompletionStream(context.Background(), chatReq)
		if err != nil {
			log.Printf("Error creating chat completion stream: %v", err)
			w.WriteString("data: error: " + err.Error() + "\n\n")
			w.Flush()
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// End of stream reached normally
					log.Printf("AI response stored successfully: %s", fullResponse)
					if fullResponse != "" {
						aiMsg := models.ChatMessage{
							PostID:  post.ID,
							Sender:  "ai",
							Content: fullResponse,
							Reason:  reasonResponse,
						}
						if errDb := models.DB.Create(&aiMsg).Error; errDb != nil {
							log.Printf("Error storing AI message: %v", err)
							// Optionally log the error; the streaming response is already complete.
						}
					}
					w.WriteString("data: [DONE]\n\n")
					w.Flush()
					return
				}
				log.Printf("Error receiving stream response: %v", err)
				w.WriteString("data: error: " + err.Error() + "\n\n")
				w.Flush()
				break
			}
			// Get the delta text from the response.

			jsonData, err := json.Marshal(map[string]interface{}{
				"choices": response.Choices,
			})
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				return
			}
			chunkApi := string(jsonData)
			chunkReason := response.Choices[0].Delta.Reasoning
			chunk := response.Choices[0].Delta.Content
			if chunkReason != "" {
				reasonResponse += chunkReason
				log.Printf("Received Reasoning chunk: %s", chunkReason)

				w.WriteString("data: " + chunkApi + "\n\n")
				w.Flush()
			}
			if chunk != "" {
				fullResponse += chunk
				log.Printf("Received chunk: %s", chunk)

				w.WriteString("data: " + chunkApi + "\n\n")
				w.Flush()
			}
		}
		// log.Printf("chunk response: %s", fullResponse)

	})

	if err != nil {
		log.Printf("Error during streaming: %v", err)
		return err
	}
	if fullResponse != "" {
		log.Printf("Full AI response: %s", fullResponse)
	}

	return nil
}
