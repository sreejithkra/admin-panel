package database

import (
	"admincontrol/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB
var err error

func Gormconnect() {
	dsn := "postgres://postgres:5102@localhost:5432/database"
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("connection failed due to ", err)
	}
	Db.AutoMigrate(&models.Signupusers{})
}
