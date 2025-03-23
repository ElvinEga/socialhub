package models

import "time"

type Post struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Content    string    `json:"content"`
	ImageUrls  []string  `json:"image_urls" gorm:"type:text[]"`
	UserID     uint      `json:"user_id"`
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	Likes      []Like    `json:"likes"`
	Comments   []Comment `json:"comments"`
	LikeCount  int64     `json:"likes" gorm:"default:0"`
	ShareCount int64     `json:"shares" gorm:"default:0"`
	ViewCount  int64     `json:"views" gorm:"default:0"`
	ILiked     bool      `json:"i_liked" gorm:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
