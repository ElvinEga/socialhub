package handlers

import (
	"socialmedia/models"
	"socialmedia/services/project"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// CreateFeature godoc
// @Summary Create a new feature
// @Description Add a new feature to an existing project
// @Tags features
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Project ID"
// @Param feature body models.Feature true "Feature details"
// @Success 201 {object} models.Feature
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/{id}/features [post]
func CreateFeature() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		projectID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
		}

		var project models.Project
		if err := models.DB.Where("id = ? AND user_id = ?", projectID, user.ID).First(&project).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Project not found"})
		}

		var feature models.Feature
		if err := c.BodyParser(&feature); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		feature.ProjectID = project.ID
		feature.UserID = user.ID

		if err := models.DB.Create(&feature).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create feature"})
		}

		return c.Status(201).JSON(feature)
	}
}

// GeneratePRD godoc
// @Summary Generate PRD for a feature
// @Description Generate a Product Requirements Document using AI
// @Tags features
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /features/{id}/generate-prd [post]
func GeneratePRD(projectService *project.Service) fiber.Handler {
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

// GetFeaturePRD godoc
// @Summary Get feature PRD
// @Description Retrieve the Product Requirements Document for a feature
// @Tags features
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Feature ID"
// @Success 200 {object} models.Prd
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /features/{id}/prd [get]
func GetFeaturePRD() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		featureID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid feature ID"})
		}

		var prd models.Prd
		if err := models.DB.Where("feature_id = ? AND user_id = ?", featureID, user.ID).First(&prd).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "PRD not found"})
		}

		return c.JSON(prd)
	}
}
