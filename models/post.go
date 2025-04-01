package models

import "time"

type Post struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Content string `json:"content"`

	Media         []Media   `json:"media" gorm:"foreignKey:PostID"`
	UserID        uint      `json:"user_id"`
	User          User      `json:"user" gorm:"foreignKey:UserID"`
	Likes         []Like    `json:"likes" gorm:"foreignKey:PostID"`
	Comments      []Comment `json:"comments" gorm:"foreignKey:PostID"`
	PostType      string    `gorm:"default:'regular'" json:"post_type"`
	CommentsCount int64     `json:"comments_count" gorm:"default:0"`
	LikeCount     int64     `json:"likes_count" gorm:"default:0"`
	ShareCount    int64     `json:"shares_count" gorm:"default:0"`
	ViewCount     int64     `json:"views_count" gorm:"default:0"`
	ILiked        bool      `json:"i_liked" gorm:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
