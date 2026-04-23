package erpnext

type CustomerPayload struct {
	Doctype       string `json:"doctype"`
	CustomerName  string `json:"customer_name"`
	CustomerType  string `json:"customer_type"`
	CustomerGroup string `json:"customer_group"` 
	Territory     string `json:"territory"`
	EmailId       string `json:"email_id"`
	MobileNo      string `json:"mobile_no"`
}