package models

type CheckoutItem struct {
	ItemCode string  `json:"item_code"`
	ItemName string  `json:"item_name"`
	Qty      float64 `json:"qty"`
	Rate     float64 `json:"rate"`
	VariantLainnya []CartVariantLainnya `json:"variant_lainnya"`
}

type CheckoutRequest struct {
	AddressName string         `json:"address_name"`
	Items       []CheckoutItem `json:"items"`
}

type SalesOrderResponse struct {
	Data struct {
		Name string `json:"name"`
	} `json:"data"`
}

type ERPNextAddressResponse struct {
	Data struct {
		AddressTitle string `json:"address_title"`
		AddressLine1 string `json:"address_line1"`
		City         string `json:"city"`
		Pincode      string `json:"pincode"`
		Phone        string `json:"phone"`
	} `json:"data"`
}
