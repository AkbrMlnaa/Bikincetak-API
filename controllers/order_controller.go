package controllers

import (
	"bikincetak-api/database"
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)


func CreateOrder(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak, token tidak ditemukan di sistem."})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Gagal membaca struktur token."})
	}

	var customerID string
	if claims["customer_id"] != nil {
		customerID = fmt.Sprintf("%v", claims["customer_id"])
	}

	if customerID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Customer ID tidak ditemukan di dalam token. Pastikan Anda Login menggunakan versi API terbaru!",
		})
	}

	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format pesanan tidak valid"})
	}

	deliveryDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	transactionDate := time.Now().Format("2006-01-02")

	var soItems []map[string]interface{}
	for _, item := range req.Items {
		soItems = append(soItems, map[string]interface{}{
			"item_code":     item.ItemCode,
			"qty":           item.Qty,
			"rate":          item.Rate,
			"delivery_date": deliveryDate,
		})
	}

	payload := map[string]interface{}{
		"customer":         customerID, 
		"transaction_date": transactionDate,
		"delivery_date":    deliveryDate,
		"items":            soItems,
		"docstatus":        0,  
	}

	payloadBytes, _ := json.Marshal(payload)

	res, errERP := erpnext.ERPNextReq("POST", "/api/resource/Sales Order", payloadBytes)
	if errERP != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghubungi server ERPNext"})
	}

	var soRes models.SalesOrderResponse
	if errUnmarshal := json.Unmarshal(res, &soRes); errUnmarshal != nil || soRes.Data.Name == "" {
		fmt.Println("[ERROR ERPNEXT]:", string(res)) 
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat pesanan di server pusat"})
	}

	redisKey := "cart:" + customerID 
	database.Rdb.Del(database.Ctx, redisKey)

	return c.Status(201).JSON(fiber.Map{
		"message":  "Pesanan berhasil dibuat!",
		"id_order": soRes.Data.Name,
		"status":   "Draft",
	})
}

// func CreateOrder(c *fiber.Ctx) error {

// 	authHeader := c.Get("Authorization")
	
// 	// Cek apakah header kosong
// 	if authHeader == "" {
// 		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak. Header Authorization kosong."})
// 	}

// 	// Cek apakah formatnya "Bearer <token>"
// 	if !strings.HasPrefix(authHeader, "Bearer ") {
// 		return c.Status(401).JSON(fiber.Map{"error": "Format token salah. Harus berawalan 'Bearer '"})
// 	}

// 	// Ambil token aslinya (buang kata "Bearer ")
// 	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 	secret := os.Getenv("JWT_SECRET")

// 	// Parsing dan Validasi Token
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(secret), nil
// 	})

// 	if err != nil || !token.Valid {
// 		fmt.Println("❌ [JWT ERROR]:", err) 
// 		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak. Token tidak valid atau sudah kedaluwarsa."})
// 	}

// 	// Ekstrak data dari dalam token
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return c.Status(401).JSON(fiber.Map{"error": "Gagal membaca struktur token"})
// 	}

// 	// Ambil customer_id (menggunakan fmt.Sprintf agar aman dari error beda tipe data)
// 	customerID := fmt.Sprintf("%v", claims["customer_id"])
// 	if customerID == "" || customerID == "<nil>" {
// 		return c.Status(400).JSON(fiber.Map{"error": "Customer ID tidak ditemukan di dalam token!"})
// 	}

// 	// ==========================================
// 	// 🛒 2. PROSES PEMBUATAN SALES ORDER
// 	// ==========================================
// 	var req models.CheckoutRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Format pesanan tidak valid"})
// 	}

// 	// Set tanggal pengiriman default (misal hari ini + 3 hari)
// 	deliveryDate := time.Now().AddDate(0, 0, 3).Format("2006-01-02")
// 	transactionDate := time.Now().Format("2006-01-02")

// 	// Mapping item dari request Frontend ke format ERPNext
// 	var soItems []map[string]interface{}
// 	for _, item := range req.Items {
// 		soItems = append(soItems, map[string]interface{}{
// 			"item_code":     item.ItemCode,
// 			"qty":           item.Qty,
// 			"rate":          item.Rate,
// 			"delivery_date": deliveryDate,
// 		})
// 	}

// 	payload := map[string]interface{}{
// 		"customer":         customerID, // Didapat dari token JWT
// 		"transaction_date": transactionDate,
// 		"delivery_date":    deliveryDate,
// 		"items":            soItems,
// 		"docstatus":        0, // 0 = Draft
// 		"naming_series":    "SO-", // Sesuaikan Naming Series di ERPNext-mu
// 	}

// 	payloadBytes, _ := json.Marshal(payload)

// 	res, errERP := erpnext.ERPNextReq("POST", "/api/resource/Sales Order", payloadBytes)
// 	if errERP != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghubungi server ERPNext"})
// 	}

// 	var soRes models.SalesOrderResponse
// 	// Cek apakah balasan sukses dengan melihat IdOrder (yang tag json-nya "name")
// 	if errUnmarshal := json.Unmarshal(res, &soRes); errUnmarshal != nil || soRes.Data.Name == "" {
// 		fmt.Println("❌ [ERROR ERPNEXT]:", string(res)) // CCTV Terminal untuk ERPNext
// 		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat pesanan di server pusat"})
// 	}

// 	// Balikkan response sukses ke Frontend
// 	return c.Status(201).JSON(fiber.Map{
// 		"message":  "Pesanan berhasil dibuat!",
// 		"id_order": soRes.Data.Name,
// 		"status":   "Draft",
// 	})
// }