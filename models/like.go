package models

import "time"

type Like struct {
	ID        uint  `gorm:"primaryKey" json:"id"`
	UserID    uint  `json:"user_id"`
	PostID    *uint `json:"post_id"`
	CommentID *uint `json:"comment_id"`

	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Post      *Post     `json:"post,omitempty" gorm:"foreignKey:PostID"`
	Comment   *Comment  `json:"comment,omitempty" gorm:"foreignKey:CommentID"`
	CreatedAt time.Time `json:"created_at"`
}
