package database

import (
	"bikincetak-api/models"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB()  {
	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal connect ke database : ", err)
	}

	log.Println("Berhasil terhubung ke database")

	err = db.AutoMigrate(
		&models.User{},
		&models.Cart{},
		&models.CartItem{},
	)
	if err != nil {
		log.Println("Gagal melakukan migrasi tabel:", err)
	}

	DB = db
}