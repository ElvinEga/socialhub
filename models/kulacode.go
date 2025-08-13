package models

import (
	"time"

	"gorm.io/gorm"
)

// Project model
// @Description Project information with AI-generated plan
type Project struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	Name        string          `json:"name" example:"E-commerce Platform"`
	Description string          `json:"description" example:"A modern e-commerce platform with AI recommendations"`
	UserID      uint            `json:"userId" example:"1"`
	User        User            `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TechStack   TechnologyStack `json:"techStack,omitempty" gorm:"foreignKey:ProjectID"`
	Features    []Feature       `json:"features,omitempty" gorm:"foreignKey:ProjectID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TechnologyStack model
// @Description Technology stack for a project
type TechnologyStack struct {
	ID          uint        `gorm:"primarykey" json:"id"`
	ProjectID   uint        `json:"projectId" example:"1"`
	UserID      uint        `json:"userId" example:"1"`
	Description string      `json:"description" example:"Modern web application stack"`
	StackItems  []StackItem `json:"stackItems,omitempty" gorm:"foreignKey:TechStackID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	User User `gorm:"foreignKey:UserID"`
}

// StackItem model
// @Description Individual technology in the stack
type StackItem struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	TechStackID uint   `json:"techStackId" example:"1"`
	UserID      uint   `json:"userId" example:"1"`
	Name        string `json:"name" example:"React"`
	Overview    string `json:"overview" example:"Frontend JavaScript library"`
	Details     string `json:"details" example:"Detailed description of React usage"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relationships
	TechStack TechnologyStack `gorm:"foreignKey:TechStackID"`
	User      User            `gorm:"foreignKey:UserID"`
}

// Feature model
// @Description Project feature with PRD
type Feature struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	ProjectID uint   `json:"projectId" example:"1"`
	UserID    uint   `json:"userId" example:"1"`
	Name      string `json:"name" example:"Feature Name"`
	Overview  string `json:"overview" example:"User login and registration system"`
	Details   string `json:"details" example:"Detailed feature description"`
	Prd       *Prd   `json:"prd,omitempty" gorm:"foreignKey:FeatureID"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Project Project `gorm:"foreignKey:ProjectID"`
	User    User    `gorm:"foreignKey:UserID"`
}

// Prd model
// @Description Product Requirements Document
type Prd struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	FeatureID uint   `json:"featureId" example:"1"`
	UserID    uint   `json:"userId" example:"1"`
	Content   string `json:"content" example:"# PRD for User Authentication\n## User Stories\nAs a user, I want to..."`
	CreatedAt time.Time
	UpdatedAt time.Time

	Feature Feature `gorm:"foreignKey:FeatureID"`
	User    User    `gorm:"foreignKey:UserID"`
}
