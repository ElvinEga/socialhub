package project

import (
	"log"
	"socialmedia/models"
	"socialmedia/services/ai"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
	ai *ai.AIService
}

func NewService(ai *ai.AIService) *Service {
	return &Service{db: models.DB, ai: ai}
}

func (s *Service) CreateProjectWithAIPlanning(project *models.Project, userID uint) error {
	project.UserID = userID

	if err := s.db.Create(project).Error; err != nil {
		return err
	}

	plan, err := s.ai.GenerateProjectPlan(project.Name, project.Description)
	if err != nil {
		log.Printf("AI planning failed: %v", err)
		return nil
	}

	techStack := models.TechnologyStack{
		ProjectID:   project.ID,
		UserID:      userID,
		Description: plan.TechStack.Description,
	}
	if err := s.db.Create(&techStack).Error; err != nil {
		return err
	}

	for _, item := range plan.TechStack.Items {
		stackItem := models.StackItem{
			TechStackID: techStack.ID,
			UserID:      userID,
			Name:        item.Name,
			Overview:    item.Overview,
			Details:     item.Details,
		}
		if err := s.db.Create(&stackItem).Error; err != nil {
			return err
		}
	}

	for _, feature := range plan.Features {
		newFeature := models.Feature{
			ProjectID: project.ID,
			UserID:    userID,
			Name:      feature.Name,
			Overview:  feature.Overview,
			Details:   feature.Details,
		}
		if err := s.db.Create(&newFeature).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GeneratePRDForFeature(featureID, userID uint) error {
	var feature models.Feature
	if err := s.db.Preload("Project").Where("id = ? AND user_id = ?", featureID, userID).First(&feature).Error; err != nil {
		return err
	}

	prdContent, err := s.ai.GeneratePRD(feature.Name, feature.Details)
	if err != nil {
		return err
	}

	prd := models.Prd{
		FeatureID: feature.ID,
		UserID:    userID,
		Content:   prdContent.Content,
	}

	if feature.Prd != nil {
		prd.ID = feature.Prd.ID
		if err := s.db.Save(&prd).Error; err != nil {
			return err
		}
	} else {
		if err := s.db.Create(&prd).Error; err != nil {
			return err
		}
	}

	return nil
}
