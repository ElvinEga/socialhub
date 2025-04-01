package models

import (
	"gorm.io/gorm"
)

type Follow struct {
	gorm.Model
	FollowerID  uint `gorm:"primaryKey" json:"follower_id"`
	FollowingID uint `gorm:"primaryKey" json:"following_id"`

	Followers User `gorm:"foreignKey:FollowerID" json:"follower"`
	Following User `gorm:"foreignKey:FollowingID" json:"following"`
}
