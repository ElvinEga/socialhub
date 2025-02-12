package main

import (
	"log"
	"socialmedia/config"
	_ "socialmedia/docs"
	"socialmedia/models"
	"socialmedia/routes"

	fiberSwagger "github.com/swaggo/fiber-swagger"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize configuration (loads .env if present)
	config.InitConfig()

	// Connect to the database and run migrations
	db := models.ConnectDatabase()
	models.Migrate(db)

	// Initialize the Fiber app
	app := fiber.New()
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Register API routes
	routes.Setup(app)

	// Start the server on port 3000
	log.Fatal(app.Listen(":8000"))
}
