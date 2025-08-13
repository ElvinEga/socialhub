package handlers

import (
	"socialmedia/models"
	"socialmedia/services/project"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateFeature(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		projectID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
		}

		var project models.Project
		if err := db.Where("id = ? AND user_id = ?", projectID, user.ID).First(&project).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Project not found"})
		}

		var feature models.Feature
		if err := c.BodyParser(&feature); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		feature.ProjectID = project.ID
		feature.UserID = user.ID

		if err := db.Create(&feature).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create feature"})
		}

		return c.Status(201).JSON(feature)
	}
}

func GeneratePRD(db *gorm.DB, projectService *project.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		featureID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid feature ID"})
		}

		if err := projectService.GeneratePRDForFeature(uint(featureID), user.ID); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate PRD"})
		}

		return c.JSON(fiber.Map{"message": "PRD generated successfully"})
	}
}

func GetFeaturePRD(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		featureID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid feature ID"})
		}

		var prd models.Prd
		if err := db.Where("feature_id = ? AND user_id = ?", featureID, user.ID).First(&prd).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "PRD not found"})
		}

		return c.JSON(prd)
	}
}
