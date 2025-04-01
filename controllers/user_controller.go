package controllers

import (
	"socialmedia/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ProfileResponse represents the response structure for the GetProfile function.
type ProfileResponse struct {
	ID             uint      `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Username       string    `json:"username"`
	ProfilePicture string    `json:"profile_picture"`
	Bio            string    `json:"bio"`
	CreatedAt      time.Time `json:"created_at"`
	FollowersCount int       `json:"followers_count"`
	FollowingCount int       `json:"following_count"`
}

// GetProfile returns the profile of the authenticated user.
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} ProfileResponse
// @Failure 404 {object} ErrorResponse
// @Router /profile [get]
// @Security ApiKeyAuth
func GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user models.User
	if err := models.DB.Preload("Followers").Preload("Following").First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "User not found"})
	}

	// Build profile response.
	profile := ProfileResponse{
		ID:             user.ID,
		Email:          user.Email,
		Name:           user.Name,
		Username:       user.Username,
		ProfilePicture: user.ProfilePicture,
		Bio:            user.Bio,
		CreatedAt:      user.CreatedAt,
		FollowersCount: len(user.Followers),
		FollowingCount: len(user.Following),
	}
	return c.JSON(profile)
}

// FollowUser lets the current user follow another user.
// @Summary Follow a user
// @Description Follow another user by their ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/follow [post]
// @Security ApiKeyAuth
func FollowUser(c *fiber.Ctx) error {
	currentUserID := c.Locals("user_id").(uint)
	targetID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid user ID"})
	}

	if uint(targetID) == currentUserID {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Cannot follow yourself"})
	}

	// Check if already following.
	var follow models.Follow
	if err := models.DB.Where("follower_id = ? AND following_id = ?", currentUserID, targetID).First(&follow).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Already following"})
	}

	follow = models.Follow{
		FollowerID:  currentUserID,
		FollowingID: uint(targetID),
	}

	if err := models.DB.Create(&follow).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(MessageResponse{Message: "User followed"})
}

// UnfollowUser lets the current user unfollow another user.
// @Summary Unfollow a user
// @Description Unfollow another user by their ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/unfollow [post]
// @Security ApiKeyAuth
func UnfollowUser(c *fiber.Ctx) error {
	currentUserID := c.Locals("user_id").(uint)
	targetID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid user ID"})
	}

	var follow models.Follow
	if err := models.DB.Where("follower_id = ? AND following_id = ?", currentUserID, targetID).First(&follow).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Not following"})
	}

	if err := models.DB.Delete(&follow).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(MessageResponse{Message: "User unfollowed"})
}
