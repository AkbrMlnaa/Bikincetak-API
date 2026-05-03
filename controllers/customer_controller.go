package controllers

import (
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)


func AddCustomerAddress(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak, token tidak valid."})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Gagal membaca token."})
	}

	customerID := fmt.Sprintf("%v", claims["customer_id"])
	if customerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Customer ID tidak ditemukan di token"})
	}


	var req models.AddAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data alamat tidak valid"})
	}

	payload := map[string]interface{}{
		"address_title":                    req.AddressTitle,
		"address_type":                     req.AddressType,
		"address_line1":                    req.AddressLine1,
		"city":                             req.City,
		"state":                            req.State,
		"pincode":                          req.Pincode,
		"country":                          req.Country,
		"phone":                            req.Phone,

		"custom_rajaongkir_city_id":        req.CityID,
		"custom_rajaongkir_province_id":    req.ProvinceID,
		"custom_rajaongkir_subdistrict_id": req.SubdistrictID, 
		"links": []map[string]interface{}{
			{
				"link_doctype": "Customer",
				"link_name":    customerID,
			},
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	res, errERP := erpnext.ERPNextReq("POST", "/api/resource/Address", payloadBytes)
	if errERP != nil {
		fmt.Println("[ERROR ERPNEXT ADDRESS]:", errERP.Error())
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan alamat ke server pusat"})
	}

	var erpRes map[string]interface{}
	json.Unmarshal(res, &erpRes)

	if erpRes["data"] == nil {
		fmt.Println("[ERP Response]:", string(res))
		return c.Status(400).JSON(fiber.Map{"error": "Gagal membuat alamat, periksa kelengkapan data"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Alamat berhasil ditambahkan!",
		"data":    erpRes["data"],
	})
}