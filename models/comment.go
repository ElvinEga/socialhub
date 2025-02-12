package models

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"type:text" json:"content"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	PostID    uint      `json:"post_id"`
	Post      Post      `json:"post"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
