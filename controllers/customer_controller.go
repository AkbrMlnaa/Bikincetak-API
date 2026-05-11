package controllers

import (
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func GetProfile(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Sesi tidak valid"})
	}
	claims := userToken.Claims.(jwt.MapClaims)
	customerID := claims["customer_id"].(string)

	profile, err := erpnext.GetCustomerProfile(customerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	filteredProfile := fiber.Map{
		"customer_id":   profile["name"],
		"customer_name": profile["customer_name"],
		"email":         profile["email_id"],
		"mobile_no":     profile["mobile_no"],
		"customer_group": profile["customer_group"],
		"territory":      profile["territory"],
		"disabled":       profile["disabled"],
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   filteredProfile,
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Sesi tidak valid"})
	}
	claims := userToken.Claims.(jwt.MapClaims)
	customerID := claims["customer_id"].(string)

	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	allowedData := make(map[string]interface{})
	if val, ok := updateData["customer_name"]; ok {
		allowedData["customer_name"] = val
	}
	if val, ok := updateData["mobile_no"]; ok {
		allowedData["mobile_no"] = val
	}

	if val, ok := updateData["customer_group"]; ok {
		allowedData["customer_group"] = val
	}

	if rawPassword, ok := updateData["new_password"].(string); ok && rawPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), 10)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal memproses password baru"})
		}
		allowedData["custom_password"] = string(hashedPassword)
	}

	if len(allowedData) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Tidak ada data valid yang diubah"})
	}

	if err := erpnext.UpdateCustomer(customerID, allowedData); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Profil berhasil diperbarui!",
	})
}

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
		"data":    erpRes["data"],
	})
}

func GetCustomerAddresses(c *fiber.Ctx) error {

	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak, token tidak valid."})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Gagal membaca token."})
	}
	customerID := fmt.Sprintf("%v", claims["customer_id"])

	filtersArray := []interface{}{
		[]interface{}{"Dynamic Link", "link_doctype", "=", "Customer"},
		[]interface{}{"Dynamic Link", "link_name", "=", customerID},
	}
	filterBytes, _ := json.Marshal(filtersArray)
	
	fieldsParam := `["name", "address_title", "address_line1", "city", "pincode", "phone"]`
	
	endpoint := `/api/resource/Address?filters=` + url.QueryEscape(string(filterBytes)) + `&fields=` + url.QueryEscape(fieldsParam)

	resAddr, err := erpnext.ERPNextReq("GET", endpoint, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil daftar alamat"})
	}

	var addrResp struct {
		Data []map[string]interface{} `json:"data"`
	}
	
	if err := json.Unmarshal(resAddr, &addrResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca format alamat dari ERPNext"})
	}

	return c.Status(200).JSON(fiber.Map{
		"data":    addrResp.Data,
	})
}

func UpdateCustomerAddress(c *fiber.Ctx) error {
	addressName := c.Params("name")
	if addressName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "ID Alamat tidak boleh kosong"})
	}

	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak, token tidak valid."})
	}
	claims := userToken.Claims.(jwt.MapClaims)
	customerID := fmt.Sprintf("%v", claims["customer_id"])

	var req models.UpdateAddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	checkEndpoint := "/api/resource/Address/" + addressName
	resCheck, errCheck := erpnext.ERPNextReq("GET", checkEndpoint, nil)
	
	if errCheck != nil || strings.Contains(string(resCheck), "exc_type") {
		return c.Status(404).JSON(fiber.Map{"error": "Alamat tidak ditemukan"})
	}

	var addressData struct {
		Data struct {
			Links []struct {
				LinkDoctype string `json:"link_doctype"`
				LinkName    string `json:"link_name"`
			} `json:"links"`
		} `json:"data"`
	}
	json.Unmarshal(resCheck, &addressData)

	isOwner := false
	for _, link := range addressData.Data.Links {
		if link.LinkDoctype == "Customer" && link.LinkName == customerID {
			isOwner = true
			break
		}
	}

	if !isOwner {
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak memiliki akses untuk mengubah alamat ini!"})
	}

	updatePayload := map[string]interface{}{}
	
	if req.AddressTitle != ""  { updatePayload["address_title"] = req.AddressTitle }
	if req.AddressType != ""   { updatePayload["address_type"] = req.AddressType }
	if req.AddressLine1 != ""  { updatePayload["address_line1"] = req.AddressLine1 }
	if req.City != ""          { updatePayload["city"] = req.City }
	if req.State != ""         { updatePayload["state"] = req.State }
	if req.Pincode != ""       { updatePayload["pincode"] = req.Pincode }
	if req.Country != ""       { updatePayload["country"] = req.Country }
	if req.Phone != ""         { updatePayload["phone"] = req.Phone }

	if req.CityID != ""        { updatePayload["custom_city_id"] = req.CityID }
	if req.ProvinceID != ""    { updatePayload["custom_province_id"] = req.ProvinceID }
	if req.SubdistrictID != "" { updatePayload["custom_subdistrict_id"] = req.SubdistrictID }

	payloadBytes, _ := json.Marshal(updatePayload)

	res, err := erpnext.ERPNextReq("PUT", checkEndpoint, payloadBytes)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghubungi server ERPNext"})
	}

	if strings.Contains(string(res), "exc_type") {
		fmt.Println("[ERPNEXT UPDATE ERROR]:", string(res))
		return c.Status(400).JSON(fiber.Map{
			"error": "Gagal memperbarui alamat. Cek kembali format data.",
			"detail": json.RawMessage(res),
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "Alamat berhasil diperbarui",
	})
}

func DeleteCustomerAddress(c *fiber.Ctx) error {
    addressName := c.Params("name")
    if addressName == "" {
        return c.Status(400).JSON(fiber.Map{"error": "ID Alamat tidak boleh kosong"})
    }

    userToken, ok := c.Locals("user").(*jwt.Token)
    if !ok || userToken == nil {
        return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak"})
    }
    claims := userToken.Claims.(jwt.MapClaims)
    customerID := fmt.Sprintf("%v", claims["customer_id"])

    checkEndpoint := "/api/resource/Address/" + addressName 
    resCheck, errCheck := erpnext.ERPNextReq("GET", checkEndpoint, nil)
    
    if errCheck != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal terhubung ke server ERPNext"})
    }

    if strings.Contains(string(resCheck), "exc_type") {
        fmt.Println("[GET ADDRESS ERROR]:", string(resCheck)) 
        return c.Status(404).JSON(fiber.Map{
            "error": "Alamat tidak ditemukan di sistem untuk dicek kepemilikannya.",
        })
    }

    var addressData struct {
        Data struct {
            Links []struct {
                LinkDoctype string `json:"link_doctype"`
                LinkName    string `json:"link_name"`
            } `json:"links"`
        } `json:"data"`
    }

    if err := json.Unmarshal(resCheck, &addressData); err != nil {
        fmt.Println("[JSON UNMARSHAL ERROR]:", err)
        return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca struktur data alamat"})
    }

    isOwner := false
    for _, link := range addressData.Data.Links {
        if link.LinkDoctype == "Customer" && link.LinkName == customerID {
            isOwner = true
            break
        }
    }

    if !isOwner {
        return c.Status(403).JSON(fiber.Map{"error": "Anda tidak memiliki akses untuk menghapus alamat ini!"})
    }

    res, err := erpnext.ERPNextReq("DELETE", checkEndpoint, nil)
    
    if err != nil || strings.Contains(string(res), "exc_type") {
        fmt.Println("[DELETE ADDRESS ERROR]:", string(res))
        return c.Status(400).JSON(fiber.Map{
            "error": "Alamat tidak bisa dihapus karena mungkin sedang terikat dengan pesanan aktif.",
        })
    }

    return c.Status(200).JSON(fiber.Map{
        "message": "Alamat berhasil dihapus",
    })
}