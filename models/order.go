package models

type CheckoutItem struct {
	ItemCode string  `json:"item_code"`
	Qty      float64 `json:"qty"`
	Rate     float64 `json:"rate"`
}

type CheckoutRequest struct {
	Items []CheckoutItem `json:"items"`
}


type SalesOrderResponse struct {
	Data struct {
		Name string `json:"name"` 
	} `json:"data"`
}