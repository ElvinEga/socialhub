package models

import "gorm.io/gorm"

type Like struct {
	gorm.Model
	UserID    uint  `json:"user_id"`
	PostID    *uint `json:"post_id"`
	CommentID *uint `json:"comment_id"`

	User    User     `json:"user" gorm:"foreignKey:UserID"`
	Post    *Post    `json:"post,omitempty" gorm:"foreignKey:PostID"`
	Comment *Comment `json:"comment,omitempty" gorm:"foreignKey:CommentID"`
}
