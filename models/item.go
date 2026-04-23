package models


type AttributeValue struct {
	Value string `json:"value"` 
}


type ItemAttribute struct {
	Attribute       string           `json:"attribute"`
	AttributeValues []AttributeValue `json:"attribute_value"`
}

type Template struct {
	Name        string          `json:"name"`
	ItemName    string          `json:"item_name"`
	HasVariants int             `json:"has_variants"`
	Attributes  []ItemAttribute `json:"attributes"`
}

type Items struct {
	Data []Template `json:"data"`
}