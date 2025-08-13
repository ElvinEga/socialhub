package models

import (
	"log"
	"socialmedia/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	DB = db
	return DB
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Post{},
		&Comment{},
		&Like{},
		&Follow{},
		&ChatMessage{},
		&Media{},
		&Project{},
		TechnologyStack{},
		&StackItem{},
		&Feature{},
		&Prd{})
}
