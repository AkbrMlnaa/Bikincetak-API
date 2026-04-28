package models

import (
	"time"

)


type Cart struct {
    ID        uint       `gorm:"primaryKey" json:"id"`
    Email     string     `gorm:"index" json:"email"` 
    CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"` 
	Items     []CartItem `gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"items"`
}


type CartItem struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	CartID      uint    `gorm:"index" json:"cart_id"`
	ItemCode    string  `json:"item_code"`    
	VariantName string  `json:"variant_name"` 
	Qty         int     `json:"qty"`
	Price       float64 `json:"price"`        
	ImageURL    string  `json:"image_url"`    
	UOM         string  `json:"uom"`          
	Notes       string  `json:"notes"`
}


type AddToCartRequest struct {
	ItemCode    string  `json:"item_code"`
	VariantName string  `json:"variant_name"`
	Qty         int     `json:"qty"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"` 
	UOM         string  `json:"uom"`       
	Notes       string  `json:"notes"`     
}

type UpdateCartRequest struct {
	Qty   int    `json:"qty"`
	Notes string `json:"notes"` 
	Price float64 `json:"price"`
}