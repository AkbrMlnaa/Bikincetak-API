package controllers

import (
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memproses keamanan password"})
	}

	customerID, err := erpnext.CreateCustomer(req.Name, req.Email, req.Number, string(hashedPassword))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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

	customerID, dbPassword, err := erpnext.GetCustomerAuthData(req.Email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Email atau Password salah",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Email atau Password salah",
		})
	}

	claims := jwt.MapClaims{
		"email":       req.Email,
		"customer_id": customerID,
		"expired":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menggenerate token",
		})
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "jwt"
	cookie.Value = t
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true 
	cookie.Secure = false  
	cookie.SameSite = "Lax"

	c.Cookie(cookie) 

	return c.JSON(fiber.Map{
		"status":  true,
		"message": "Login sukses",
	})
}

func Logout(c *fiber.Ctx) error {
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour) // Set waktu mundur agar browser otomatis menghapusnya
	cookie.HTTPOnly = true

	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"message": "Logout berhasil",
	})
}