package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	Content       string `json:"content"`
	UserID        uint   `json:"user_id"`
	PostType      string `gorm:"default:'regular'" json:"post_type"`
	CommentsCount int64  `json:"comments_count" gorm:"default:0"`
	LikeCount     int64  `json:"likes_count" gorm:"default:0"`
	ShareCount    int64  `json:"shares_count" gorm:"default:0"`
	ViewCount     int64  `json:"views_count" gorm:"default:0"`
	ILiked        bool   `json:"i_liked" gorm:"-"`

	User     User      `json:"user" gorm:"foreignKey:UserID"`
	Media    []Media   `json:"media" gorm:"foreignKey:PostID"`
	Likes    []Like    `json:"likes" gorm:"foreignKey:PostID"`
	Comments []Comment `json:"comments" gorm:"foreignKey:PostID"`
}
