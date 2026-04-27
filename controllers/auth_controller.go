package controllers

import (
	"bikincetak-api/database"
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
	var req models.Register
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Email sudah terdaftar di sistem!"})
	}

	customerID, err := erpnext.CreateCustomer(req.Name, req.Email, req.Number)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memproses keamanan password"})
	}

	newUser := models.User{
		Email:      req.Email,
		Password:   string(hashedPassword),
		CustomerId: customerID,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan user ke database lokal"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":     "Registrasi berhasil!",
		"customer_id": customerID,
	})
}

func Login(c *fiber.Ctx) error {
	var req models.Login
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "format tidak valid",
		})
	}

	var user models.User
	if err := database.DB.Where("email =?", req.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Email atau Password salah",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Email atau Password salah",
		})
	}

	claims := jwt.MapClaims{
		"email":       user.Email,
		"customer_id": user.CustomerId,
		"expired":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "gagal menggenerate token",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"token":  t,
	})
}
