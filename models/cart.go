package models

import "time"


type Cart struct {
    ID        uint       `gorm:"primaryKey" json:"id"`
    Email     string     `gorm:"index" json:"email"` 
    CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"` 
    Items     []CartItem `gorm:"foreignKey:CartID" json:"items"`
}


type CartItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	CartID      uint    `gorm:"index" json:"cart_id"`
	ItemCode    string  `json:"item_code"`    
	VariantName string  `json:"variant_name"` 
	Qty         int     `json:"qty"`
	Price       float64 `json:"price"`        
}


type AddToCartRequest struct {
	ItemCode    string  `json:"item_code"`
	VariantName string  `json:"variant_name"`
	Qty         int     `json:"qty"`
	Price       float64 `json:"price"`
}