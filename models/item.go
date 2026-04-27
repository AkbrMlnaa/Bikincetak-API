package models


type ItemGroup struct {
	ItemGroupName string          `json:"item_group_name"`
	Templates     []*ItemTemplate `json:"templates"`
}

type ItemTemplate struct {
	Name       string          `json:"name"`
	ItemName   string          `json:"item_name"`
	ImageURL   string          `json:"image_url"`
	Attributes []ItemAttribute `json:"attributes"`
}

type ItemAttribute struct {
	Attribute       string           `json:"attribute"`
	AttributeValues []AttributeValue `json:"attribute_values"`
}

type AttributeValue struct {
	AttributeValue string `json:"attribute_value"`
}





type ItemVariant struct {
	VariantName  string        `json:"variant_name"`
	ItemCode     string        `json:"item_code"`
	UOM          string        `json:"uom"`
	Description  string        `json:"description"`
	PricingRules []PricingRule `json:"pricing_rules"`
}


type PricingRule struct {
	MinQty float64 `json:"min_qty"`
	MaxQty float64 `json:"max_qty"`
	Price  float64 `json:"price"`
}