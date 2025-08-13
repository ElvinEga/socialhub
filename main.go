package main

import (
	"log"
	"socialmedia/config"
	_ "socialmedia/docs"
	"socialmedia/models"
	"socialmedia/routes"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Initialize configuration (loads .env if present)
	config.InitConfig()

	// Connect to the database and run migrations
	db := models.ConnectDatabase()
	models.Migrate(db)

	// Initialize the Fiber app
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,https://tours-dashboard-pi.vercel.app", // or your Next.js URL
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Register API routes
	routes.Setup(app)

	// Start the server on port 3000
	log.Fatal(app.Listen(":8000"))
}
