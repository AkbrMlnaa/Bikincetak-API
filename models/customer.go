package models

type AddAddressRequest struct {
	AddressTitle  string `json:"address_title"`
	AddressType   string `json:"address_type"`
	AddressLine1  string `json:"address_line1"`
	AddressLine2  string `json:"address_line2"` 
	City          string `json:"city"`          
	State         string `json:"state"`         
	Pincode       string `json:"pincode"`
	Country       string `json:"country"`
	Phone         string `json:"phone"`
	
	CityID        string `json:"city_id"`        
	ProvinceID    string `json:"province_id"`    
	SubdistrictID string `json:"subdistrict_id"` 
}

type UpdateAddressRequest struct {
	AddressLine1 string `json:"address_line1"`
	City         string `json:"city"`
	Phone        string `json:"phone"`
	Postcode     string `json:"postcode"`
}