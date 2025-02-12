package models

import (
	"time"
)

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Email          string    `gorm:"unique;not null" json:"email"`
	Password       string    `json:"-"` // omit in JSON responses
	ProfilePicture string    `json:"profile_picture"`
	Name           string    `json:"name"`
	Username       string    `gorm:"unique" json:"username"`
	Bio            string    `json:"bio"`
	CreatedAt      time.Time `json:"created_at"`

	Posts []Post `json:"posts"`
	// Self-referential many-to-many for followers and following.
	Followers []*User `gorm:"many2many:follows;joinForeignKey:FollowingID;joinReferences:FollowerID" json:"followers"`
	Following []*User `gorm:"many2many:follows;joinForeignKey:FollowerID;joinReferences:FollowingID" json:"following"`
}
