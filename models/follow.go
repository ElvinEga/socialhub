package models

import "time"

// Follow represents the join table for user follow relationships.
type Follow struct {
	FollowerID  uint      `gorm:"primaryKey" json:"follower_id"`
	FollowingID uint      `gorm:"primaryKey" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}
