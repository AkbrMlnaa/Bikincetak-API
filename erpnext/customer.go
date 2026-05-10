package erpnext

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

func CreateCustomer(name, email, number, hashedPassword string) (string, error) {

	payload := CustomerPayload{
		Doctype:       "Customer",
		CustomerName:  name,
		EmailId:       email,
		CustomerType:  "Individual",
		CustomerGroup: "Individual",
		Territory:     "Indonesia",
		MobileNo:      number,
		CustomPassword: hashedPassword,
	}

	payloadBytes, _ := json.Marshal(payload)

	resp, err := ERPNextReq("POST", "/api/resource/Customer", payloadBytes)
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


func GetCustomerAuthData(email string) (string, string, error) {
	filterStr := fmt.Sprintf(`[["email_id","=","%s"]]`, email)
	encodedFilter := url.QueryEscape(filterStr)
	
	endpoint := fmt.Sprintf(`/api/resource/Customer?fields=["name","custom_password"]&filters=%s`, encodedFilter)
	
	resp, err := ERPNextReq("GET", endpoint, nil)
	if err != nil {
		return "", "", errors.New("gagal mengambil data pelanggan")
	}

	var erpResp struct {
		Data []struct {
			Name           string `json:"name"`
			CustomPassword string `json:"custom_password"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(resp, &erpResp); err != nil || len(erpResp.Data) == 0 {
		return "", "", errors.New("email tidak ditemukan di data customer")
	}

	return erpResp.Data[0].Name, erpResp.Data[0].CustomPassword, nil
}


func GetCustomerProfile(customerID string) (map[string]interface{}, error) {
	endpoint := "/api/resource/Customer/" + customerID

	resp, err := ERPNextReq("GET", endpoint, nil)
	if err != nil {
		return nil, errors.New("gagal menghubungi server ERPNext")
	}

	var erpResp map[string]interface{}
	if err := json.Unmarshal(resp, &erpResp); err != nil {
		return nil, errors.New("gagal memproses data dari ERPNext")
	}

	data, ok := erpResp["data"].(map[string]interface{})
	if !ok {
		if _, exists := erpResp["exc"]; exists {
			return nil, errors.New("customer tidak ditemukan di ERPNext")
		}
		return nil, errors.New("format response data tidak valid")
	}

	return data, nil
}

func UpdateCustomer(customerID string, data map[string]interface{}) error {
	payloadBytes, _ := json.Marshal(data)

	endpoint := "/api/resource/Customer/" + customerID
	resp, err := ERPNextReq("PUT", endpoint, payloadBytes)
	if err != nil {
		return errors.New("gagal menghubungi server ERPNext")
	}

	var erpResp map[string]interface{}
	json.Unmarshal(resp, &erpResp)

	if _, exists := erpResp["exc"]; exists {
		return errors.New("gagal memperbarui data di ERPNext")
	}

	return nil
}