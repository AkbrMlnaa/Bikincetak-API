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
	// PERBAIKAN 2: Hapus Preload("Items") karena kita tidak membutuhkannya di sini (Biar query lebih enteng)
	if err := database.DB.Where("email = ?", email).First(&cart).Error; err != nil {
		// Keranjang belum ada, buat baru
		cart = models.Cart{Email: email}
		database.DB.Create(&cart)
	}

	var existingItem models.CartItem
	// PERBAIKAN 1: Tambahkan pengecekan 'notes' di dalam Where. 
	// Jika notes-nya beda, biarkan dia membuat row baru di blok 'else'.
	err := database.DB.Where("cart_id = ? AND item_code = ? AND notes = ?", cart.ID, req.ItemCode, req.Notes).First(&existingItem).Error

	if err == nil {
		// Jika barang SAMA dan catatan (notes) SAMA persis, gabungkan jumlahnya
		existingItem.Qty += req.Qty
		existingItem.Price = req.Price // Update harga ke yang terbaru
		database.DB.Save(&existingItem)
	} else {
		// Jika barang beda, ATAU barang sama tapi catatannya BEDA, buat baris baru
		newItem := models.CartItem{
			CartID:      cart.ID,
			ItemCode:    req.ItemCode,
			VariantName: req.VariantName,
			Qty:         req.Qty,
			Price:       req.Price,
			ImageURL:    req.ImageURL,
			UOM:         req.UOM,
			Notes:       req.Notes,
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
		return c.JSON(fiber.Map{
			"message": "Keranjang kosong",
			"data": fiber.Map{
				"items": []models.CartItem{},
				"total": 0,
			},
		})
	}

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



func UpdateCartItem(c *fiber.Ctx) error {
	email := getEmailFromToken(c)
	itemID := c.Params("id") // ID dari tabel CartItem

	var req models.UpdateCartRequest // Pastikan pakai struct yang baru
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if req.Qty <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Kuantitas minimal 1. Gunakan tombol hapus jika ingin menghapus barang."})
	}

	var cart models.Cart
	if err := database.DB.Where("email = ?", email).First(&cart).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Keranjang tidak ditemukan"})
	}

	var cartItem models.CartItem
	if err := database.DB.Where("id = ? AND cart_id = ?", itemID, cart.ID).First(&cartItem).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Barang tidak ditemukan di keranjang"})
	}

	// Terapkan perubahan
	cartItem.Qty = req.Qty
	cartItem.Notes = req.Notes 
	
	// PERBAIKAN FATAL: Update harganya menyesuaikan dengan tier grosir yang baru!
	if req.Price > 0 {
		cartItem.Price = req.Price 
	}

	database.DB.Save(&cartItem)

	return c.JSON(fiber.Map{
		"message": "Keranjang berhasil diupdate",
		"data":    cartItem,
	})
}

func DeleteCartItem(c *fiber.Ctx) error {
	email := getEmailFromToken(c)
	itemID := c.Params("id")

	var cart models.Cart
	if err := database.DB.Where("email = ?", email).First(&cart).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Keranjang tidak ditemukan"})
	}

	// Hapus item yang sesuai dengan ID dan CartID
	result := database.DB.Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&models.CartItem{})
	
	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Barang tidak ditemukan di keranjang"})
	}

	return c.JSON(fiber.Map{
		"message": "Barang berhasil dihapus dari keranjang",
	})
}