package erpnext

import (
	"encoding/json"
	"errors"
	"fmt"
)

func CreateCustomer(name, email, number string) (string, error) {

	payload := CustomerPayload{
		Doctype:       "Customer",
		CustomerName:  name,
		EmailId:       email,
		CustomerType:  "Individual",
		CustomerGroup: "Individual",
		Territory:     "Indonesia",
		MobileNo:      number,
	}

	payloadBytes, _ := json.Marshal(payload)

	resp, err := ERPNextReq("POST", "api/resource/Customer", payloadBytes)
	if err != nil {
		fmt.Println("KONEKSI GAGAL:", err)
		return "", errors.New("Gagal nembak ke server ERPNext")
	}

	var erpResp map[string]interface{}
	json.Unmarshal(resp, &erpResp)

	if _, exists := erpResp["exc"]; exists {
		return "", errors.New("gagal membuat Customer di ERPNext. Nama atau Email mungkin sudah dipakai")
	}

	customerData := erpResp["data"].(map[string]interface{})
	customerID := customerData["name"].(string)

	return customerID, nil
}
