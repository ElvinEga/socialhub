package models

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Content         string `gorm:"type:text" json:"content"`
	UserID          uint   `json:"user_id"`
	ParentCommentID *uint  `json:"parent_id,omitempty"`
	PostID          uint   `json:"post_id"`
	LikeCount       int64  `json:"likes_count" gorm:"default:0"`

	User          User      `json:"user" gorm:"foreignKey:UserID"`
	Post          Post      `json:"post" gorm:"foreignKey:PostID"`
	ParentComment *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentCommentID"`
	Replies       []Comment `gorm:"foreignKey:ParentCommentID" json:"replies,omitempty"`
	Likes         []Like    `json:"likes" gorm:"foreignKey:CommentID"`
	IsLiked       bool      `json:"is_liked" gorm:"-"`
}
