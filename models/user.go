package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email          string `gorm:"unique;not null" json:"email"`
	Password       string `json:"-"` // omit in JSON responses
	ProfilePicture string `json:"profile_picture"`
	Name           string `json:"name"`
	Username       string `gorm:"unique" json:"username"`
	Bio            string `json:"bio"`
	FollowerCount  int64  `json:"follower_count" gorm:"default:0"`
	FollowingCount int64  `json:"following_count" gorm:"default:0"`

	Posts    []Post    `json:"posts" gorm:"foreignKey:UserID"`
	Comments []Comment `json:"comments" gorm:"foreignKey:UserID"`
	Likes    []Like    `json:"likes" gorm:"foreignKey:UserID"`
	// Self-referential many-to-many for followers and following.
	Followers []*User `gorm:"many2many:follows;joinForeignKey:FollowingID;joinReferences:FollowerID" json:"followers"`
	Following []*User `gorm:"many2many:follows;joinForeignKey:FollowerID;joinReferences:FollowingID" json:"following"`
}
