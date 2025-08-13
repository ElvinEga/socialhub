package routes

import (
	"socialmedia/config"
	"socialmedia/controllers"
	"socialmedia/handlers"
	"socialmedia/middlewares"
	"socialmedia/services/ai"
	"socialmedia/services/project"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	aiService := ai.NewAIService(config.OpenRouterAPIKey)
	projectService := project.NewService(aiService)
	api := app.Group("/api")

	// Public routes.
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)
	api.Get("/auth/google", controllers.GoogleLogin)
	api.Get("/auth/google/callback", controllers.GoogleCallback)
	api.Post("/logout", controllers.Logout)

	// Protected routes (require JWT authentication).
	api.Use(middlewares.JWTMiddleware)

	// User routes.
	api.Get("/profile", controllers.GetProfile)
	api.Post("/follow/:id", controllers.FollowUser)
	api.Post("/unfollow/:id", controllers.UnfollowUser)

	// Post routes.
	api.Get("/posts", controllers.PostList)
	api.Post("/posts", controllers.CreatePost)
	api.Put("/posts/:id", controllers.EditPost)
	api.Delete("/posts/:id", controllers.DeletePost)
	api.Get("/timeline", controllers.Timeline)

	// Comment routes.
	api.Get("/posts/:id/comments", controllers.GetCommentsByPostID)
	api.Get("/comments/:id", controllers.GetCommentByID)
	api.Post("/posts/:id/comments", controllers.AddComment)
	api.Put("/comments/:id", controllers.EditComment)
	api.Delete("/comments/:id", controllers.DeleteComment)
	api.Post("/comments/:id/replies", controllers.AddReply)

	// Like routes.
	api.Post("/posts/:id/like", controllers.LikePost)
	api.Delete("/posts/:id/like", controllers.UnlikePost)

	// AI Chat Post routes.
	api.Post("/ai-posts", controllers.CreateAIChatPost)
	// api.Post("/ai-posts/:id/messages", controllers.AddChatMessage)
	api.Post("/ai-posts/:id/messages", controllers.SendAIChatMessage)
	api.Get("/ai-posts/:id", controllers.GetAIChatPost)

	// Protected routes
	protected := api.Group("/agent")

	// Project routes
	protected.Post("/projects", handlers.CreateProject(projectService))
	protected.Get("/projects", handlers.GetProjects())
	protected.Get("/projects/:id", handlers.GetProject())

	// Feature routes
	protected.Post("/projects/:id/features", handlers.CreateFeature())
	protected.Post("/features/:id/generate-prd", handlers.GeneratePRD(projectService))
	protected.Get("/features/:id/prd", handlers.GetFeaturePRD())
}
