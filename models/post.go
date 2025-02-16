package models

import (
	"time"
)

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"type:text" json:"content"`
	ImageURL  string    `json:"image_url"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	Likes     []Like    `json:"likes"`
	Comments  []Comment `json:"comments"`
	PostType  string    `gorm:"default:'regular'" json:"post_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
