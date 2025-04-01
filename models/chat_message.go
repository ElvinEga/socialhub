package models

import (
	"gorm.io/gorm"
)

type ChatMessage struct {
	gorm.Model
	PostID  uint   `json:"post_id"`
	Sender  string `json:"sender"`
	Content string `gorm:"type:text" json:"content"`
	Reason  string `gorm:"type:text" json:"reason"`
}
