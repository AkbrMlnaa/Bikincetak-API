package controllers

import (
	"bikincetak-api/config"
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

func CreateOrder(c *fiber.Ctx) error {
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request tidak valid"})
	}

	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok || userToken == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Akses ditolak, token tidak valid."})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Gagal membaca token."})
	}
	customerEmail := fmt.Sprintf("%v", claims["email"])
	customerID := fmt.Sprintf("%v", claims["customer_id"])

	var grossAmount int64 = 0
	var midtransItems []midtrans.ItemDetails
	var erpItems []map[string]interface{}

	for _, item := range req.Items {
		grossAmount += int64(item.Rate) * int64(item.Qty)

		midtransItems = append(midtransItems, midtrans.ItemDetails{
			ID:    item.ItemCode,
			Name:  item.ItemName,
			Price: int64(item.Rate),
			Qty:   int32(item.Qty),
		})

		erpItems = append(erpItems, map[string]interface{}{
			"item_code": item.ItemCode,
			"qty":       item.Qty,
			"rate":      item.Rate,
		})
	}

	var wg sync.WaitGroup
	var erpAddress models.ERPNextAddressResponse
	var customerName string
	var errFetch error

	wg.Add(2)

	go func() {
		defer wg.Done()
		resAddress, errAddr := erpnext.ERPNextReq("GET", "/api/resource/Address/"+req.AddressName, nil)
		if errAddr != nil {
			errFetch = errAddr
			return
		}
		json.Unmarshal(resAddress, &erpAddress)
	}()

	go func() {
		defer wg.Done()
		resCust, errCust := erpnext.ERPNextReq("GET", "/api/resource/Customer/"+customerID, nil)
		if errCust != nil {
			errFetch = errCust
			return
		}

		var custData map[string]interface{}
		if err := json.Unmarshal(resCust, &custData); err == nil {
			if data, ok := custData["data"].(map[string]interface{}); ok {
				if name, ok := data["customer_name"].(string); ok {
					customerName = name
				}
			}
		}
	}()

	wg.Wait() 

	if errFetch != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data alamat/customer dari sistem"})
	}

	addr := erpAddress.Data
	if customerName == "" {
		customerName = customerID 
	}

	deliveryDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	soPayload := map[string]interface{}{
		"customer":         customerID,
		"items":            erpItems,
		"customer_address": req.AddressName,
		"delivery_date":    deliveryDate,
	}

	soPayloadBytes, _ := json.Marshal(soPayload)
	resSO, errSO := erpnext.ERPNextReq("POST", "/api/resource/Sales Order", soPayloadBytes)

	if errSO != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat pesanan di sistem ERP"})
	}

	if strings.Contains(string(resSO), "exc_type") {
		return c.Status(400).JSON(fiber.Map{"error": "ERPNext menolak pesanan. Cek kelengkapan data."})
	}

	var soRes models.SalesOrderResponse
	if err := json.Unmarshal(resSO, &soRes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca ID Pesanan dari ERPNext"})
	}

	orderID := soRes.Data.Name

	billAddress := &midtrans.CustomerAddress{
		FName:       customerName,
		Phone:       addr.Phone,
		Address:     addr.AddressLine1,
		City:        addr.City,
		Postcode:    addr.Pincode,
		CountryCode: "IDN",
	}

	shipAddress := &midtrans.CustomerAddress{
		FName:       addr.AddressTitle,
		Phone:       addr.Phone,
		Address:     addr.AddressLine1,
		City:        addr.City,
		Postcode:    addr.Pincode,
		CountryCode: "IDN",
	}

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: grossAmount,
		},
		Items: &midtransItems,
		CustomerDetail: &midtrans.CustomerDetails{
			FName:    customerName,
			Email:    customerEmail,
			Phone:    addr.Phone,
			BillAddr: billAddress, 
			ShipAddr: shipAddress, 
		},
	}

	snapResp, errMidtrans := config.SnapClient.CreateTransaction(snapReq)
	if errMidtrans != nil {
		
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memproses pembayaran"})
	}

	return c.Status(200).JSON(fiber.Map{
		"message":     "Pesanan berhasil dibuat",
		"order_id":    orderID,
		"payment_url": snapResp.RedirectURL,
		"snap_token":  snapResp.Token,
	})
}

func PaymentCallback(c *fiber.Ctx) error {
	var notificationPayload map[string]interface{}

	if err := c.BodyParser(&notificationPayload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	payloadJSON, _ := json.MarshalIndent(notificationPayload, "", "  ")
	fmt.Println(string(payloadJSON))

	return c.SendStatus(200)
}
