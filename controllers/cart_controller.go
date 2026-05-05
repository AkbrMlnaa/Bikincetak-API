package controllers

import (
	"bikincetak-api/database"
	"bikincetak-api/models"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)


func getEmailFromToken(c *fiber.Ctx) string {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	return claims["email"].(string)
}


// func AddToCart(c *fiber.Ctx) error {
// 	email := getEmailFromToken(c)

// 	var req models.AddToCartRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
// 	}

// 	if req.Qty <= 0 {
// 		return c.Status(400).JSON(fiber.Map{"error": "Kuantitas minimal 1"})
// 	}

// 	var cart models.Cart
// 	if err := database.DB.Where("email = ?", email).First(&cart).Error; err != nil {
// 		// Keranjang belum ada, buat baru
// 		cart = models.Cart{Email: email}
// 		database.DB.Create(&cart)
// 	}

// 	var existingItem models.CartItem
// 	err := database.DB.Where("cart_id = ? AND item_code = ? AND notes = ?", cart.ID, req.ItemCode, req.Notes).First(&existingItem).Error

// 	if err == nil {
// 		existingItem.Qty += req.Qty
// 		existingItem.Price = req.Price 
// 		database.DB.Save(&existingItem)
// 	} else {
// 		newItem := models.CartItem{
// 			CartID:      cart.ID,
// 			ItemCode:    req.ItemCode,
// 			VariantName: req.VariantName,
// 			Qty:         req.Qty,
// 			Price:       req.Price,
// 			ImageURL:    req.ImageURL,
// 			UOM:         req.UOM,
// 			Notes:       req.Notes,
// 		}
// 		database.DB.Create(&newItem)
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "Barang berhasil ditambahkan ke keranjang!",
// 	})
// }

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
	if err := database.DB.Where("email = ?", email).First(&cart).Error; err != nil {
		// Keranjang belum ada, buat baru
		cart = models.Cart{Email: email}
		database.DB.Create(&cart)
	}

	var existingItems []models.CartItem
	database.DB.Where("cart_id = ? AND item_code = ? AND notes = ?", cart.ID, req.ItemCode, req.Notes).Find(&existingItems)

	var matchedItem *models.CartItem
	
	reqVariantsJSON, _ := json.Marshal(req.VariantLainnya) 

	for i, item := range existingItems {
		dbVariantsJSON, _ := json.Marshal(item.VariantLainnya)
		
		if string(reqVariantsJSON) == string(dbVariantsJSON) {
			matchedItem = &existingItems[i]
			break
		}
	}

	if matchedItem != nil {
		matchedItem.Qty += req.Qty
		matchedItem.Price = req.Price 
		database.DB.Save(matchedItem)
	} else {
		newItem := models.CartItem{
			CartID:         cart.ID,
			ItemCode:       req.ItemCode,
			VariantName:    req.VariantName,
			Qty:            req.Qty,
			Price:          req.Price,
			ImageURL:       req.ImageURL,
			UOM:            req.UOM,
			Notes:          req.Notes,
			VariantLainnya: req.VariantLainnya, 
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

	// Kalkulasi ulang total belanja dengan memasukkan harga Variant Lainnya
	for _, item := range cart.Items {
		// 1. Ambil harga dasar produk
		hargaItem := item.Price

		// 2. Tambahkan semua harga Variant Lainnya (Add-ons) ke harga dasar
		for _, addon := range item.VariantLainnya {
			hargaItem += addon.Price
		}

		// 3. Kalikan (Harga Dasar + Total Add-on) dengan Qty pesanan
		grandTotal += (hargaItem * float64(item.Qty))
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
	itemID := c.Params("id") 

	var req models.UpdateCartRequest 
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

	cartItem.Qty = req.Qty
	cartItem.Notes = req.Notes 
	
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

	result := database.DB.Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&models.CartItem{})
	
	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Barang tidak ditemukan di keranjang"})
	}

	return c.JSON(fiber.Map{
		"message": "Barang berhasil dihapus dari keranjang",
	})
}