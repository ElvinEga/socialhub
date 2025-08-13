package models

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name        string          `gorm:"not null"`
	Description string          `gorm:"type:text"`
	UserID      uint            `gorm:"not null;index"`
	User        User            `gorm:"foreignKey:UserID"`
	TechStack   TechnologyStack `gorm:"foreignKey:ProjectID"`
	Features    []Feature       `gorm:"foreignKey:ProjectID"`
}

type TechnologyStack struct {
	gorm.Model
	ProjectID   uint        `gorm:"not null;index"`
	UserID      uint        `gorm:"not null;index"`
	Description string      `gorm:"type:text"`
	StackItems  []StackItem `gorm:"foreignKey:TechStackID"`
}

type StackItem struct {
	gorm.Model
	TechStackID uint   `gorm:"not null;index"`
	UserID      uint   `gorm:"not null;index"`
	Name        string `gorm:"not null"`
	Overview    string `gorm:"type:text"`
	Details     string `gorm:"type:text"`
}

type Feature struct {
	gorm.Model
	ProjectID uint   `gorm:"not null;index"`
	UserID    uint   `gorm:"not null;index"`
	Name      string `gorm:"not null"`
	Overview  string `gorm:"type:text"`
	Details   string `gorm:"type:text"`
	Prd       *Prd   `gorm:"foreignKey:FeatureID"`
}

type Prd struct {
	gorm.Model
	FeatureID uint   `gorm:"not null;unique;index"`
	UserID    uint   `gorm:"not null;index"`
	Content   string `gorm:"type:text"`
}
