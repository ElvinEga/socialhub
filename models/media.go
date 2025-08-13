package models

import (
	"time"
)

type MediaType string

const (
	ImageType MediaType = "image"
	VideoType MediaType = "video"
	GifType   MediaType = "gif"
)

// Media model
// @Description Media represents an image or video attached to a post
type Media struct {
	ID        uint      `gorm:"primarykey"`
	PostID    uint      `gorm:"index;not null"`             // Foreign key to Post table
	UserID    uint      `gorm:"index;not null"`             // User who uploaded (might be same as Post.UserID)
	URL       string    `gorm:"type:varchar(255);not null"` // URL of the media (e.g., S3 link)
	Type      MediaType `gorm:"type:varchar(20);not null"`  // 'image', 'video', etc.
	AltText   string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Accessibility text

	// Relationships (optional, PostID/UserID are the main links)
	// Post Post `gorm:"foreignKey:PostID"`
	// User User `gorm:"foreignKey:UserID"`
}
