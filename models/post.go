package models

import "time"

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"`
	ImageUrls []string  `json:"image_urls" gorm:"type:text[]"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Likes     int64     `json:"likes" gorm:"default:0"`
	Shares    int64     `json:"shares" gorm:"default:0"`
	Views     int64     `json:"views" gorm:"default:0"`
	ILiked    bool      `json:"i_liked" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
