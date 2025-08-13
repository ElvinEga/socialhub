package handlers

import (
	"socialmedia/models"
	"socialmedia/services/project"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateProject(db *gorm.DB, projectService *project.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)

		var project models.Project
		if err := c.BodyParser(&project); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if err := projectService.CreateProjectWithAIPlanning(&project, user.ID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create project"})
		}

		return c.Status(201).JSON(project)
	}
}

func GetProjects(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)

		var projects []models.Project
		if err := db.Where("user_id = ?", user.ID).Preload("Features").Find(&projects).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch projects"})
		}

		return c.JSON(projects)
	}
}

func GetProject(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
		}

		var project models.Project
		if err := db.Preload("TechStack.StackItems").Preload("Features.Prd").
			Where("id = ? AND user_id = ?", id, user.ID).First(&project).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Project not found"})
		}

		return c.JSON(project)
	}
}
