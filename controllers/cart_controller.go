package controllers

import (
	"bikincetak-api/database"
	"bikincetak-api/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)


func getEmailFromToken(c *fiber.Ctx) string {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	return claims["email"].(string)
}


func AddToCart(c *fiber.Ctx) error {
	email := getEmailFromToken(c)

	var req models.AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if req.Qty <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Kuantitas minimal 1"})
	}

	var cart models.Cart
	if err := database.DB.Preload("Items").Where("email = ?", email).First(&cart).Error; err != nil {
		// Keranjang belum ada, buat baru
		cart = models.Cart{Email: email}
		database.DB.Create(&cart)
	}

	var existingItem models.CartItem
	err := database.DB.Where("cart_id = ? AND item_code = ?", cart.ID, req.ItemCode).First(&existingItem).Error

	if err == nil {
		existingItem.Qty += req.Qty
		existingItem.Price = req.Price 
		database.DB.Save(&existingItem)
	} else {
		newItem := models.CartItem{
			CartID:      cart.ID,
			ItemCode:    req.ItemCode,
			VariantName: req.VariantName,
			Qty:         req.Qty,
			Price:       req.Price,
		}
		database.DB.Create(&newItem)
	}

	return c.JSON(fiber.Map{
		"message": "Barang berhasil ditambahkan ke keranjang!",
	})
}

func GetCart(c *fiber.Ctx) error {
	email := getEmailFromToken(c)

	var cart models.Cart
	if err := database.DB.Preload("Items").Where("email = ?", email).First(&cart).Error; err != nil {
		// Jika belum punya keranjang, kembalikan data kosong yang rapi
		return c.JSON(fiber.Map{
			"message": "Keranjang kosong",
			"data": fiber.Map{
				"items": []models.CartItem{},
				"total": 0,
			},
		})
	}

	// Hitung total harga
	var grandTotal float64
	for _, item := range cart.Items {
		grandTotal += (float64(item.Qty) * item.Price)
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil memuat keranjang",
		"data": fiber.Map{
			"items": cart.Items,
			"total": grandTotal,
		},
	})
}