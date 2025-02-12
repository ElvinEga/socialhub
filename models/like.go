package models

import "time"

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	PostID    uint      `json:"post_id"`
	Post      Post      `json:"post"`
	CreatedAt time.Time `json:"created_at"`
}
