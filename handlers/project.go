package handlers

import (
	"socialmedia/models"
	"socialmedia/services/project"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with AI-generated plan
// @Tags projects
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param project body models.Project true "Project details"
// @Success 201 {object} models.Project
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects [post]
func CreateProject(projectService *project.Service) fiber.Handler {
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

// GetProjects godoc
// @Summary Get user's projects
// @Description Retrieve all projects for the authenticated user
// @Tags projects
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Project
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects [get]
func GetProjects() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)

		var projects []models.Project
		if err := models.DB.Where("user_id = ?", user.ID).Preload("Features").Find(&projects).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch projects"})
		}

		return c.JSON(projects)
	}
}

// GetProject godoc
// @Summary Get project details
// @Description Get detailed information about a specific project
// @Tags projects
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Project ID"
// @Success 200 {object} models.Project
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/{id} [get]
func GetProject() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(models.User)
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid project ID"})
		}

		var project models.Project
		if err := models.DB.Preload("TechStack.StackItems").Preload("Features.Prd").
			Where("id = ? AND user_id = ?", id, user.ID).First(&project).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Project not found"})
		}

		return c.JSON(project)
	}
}
