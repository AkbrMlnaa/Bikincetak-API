package controllers

import (
	"bikincetak-api/erpnext"
	"bikincetak-api/models"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// DTO RAW DATA
// ==========================================
type RawItemsResponse struct {
	Data []struct {
		Name        string `json:"name"`
		ItemName    string `json:"item_name"`
		ItemGroup   string `json:"item_group"`
		Image       string `json:"image"`
		HasVariants int    `json:"has_variants"`
		VariantOf   string `json:"variant_of"`
	} `json:"data"`
}

type RawItemDetailResponse struct {
	Data struct {
		Attributes []struct {
			Attribute string `json:"attribute"`
		} `json:"attributes"`
	} `json:"data"`
}

// Tambahan DTO untuk mencari Item Code berdasarkan Item Name
type RawTemplateSearchResponse struct {
	Data []struct {
		Name string `json:"name"`
	} `json:"data"`
}

type RawVariantResponse struct {
	Data []struct {
		Name        string `json:"name"`
		ItemName    string `json:"item_name"`
		StockUOM    string `json:"stock_uom"`
		Description string `json:"description"`
	} `json:"data"`
}


type RawPricingRuleResponse struct {
	Data []struct {
		ItemCode string  `json:"item_code"`
		MinQty   float64 `json:"min_qty"`
		MaxQty   float64 `json:"max_qty"`
		Price    float64 `json:"price"`
	} `json:"data"`
}



func GetItems(c *fiber.Ctx) error {
	fieldsParam := `["name","item_name","item_group","image","has_variants","variant_of"]`
	itemEndpoint := `/api/resource/Item?fields=` + url.QueryEscape(fieldsParam) + `&limit=1000`

	itemRes, err := erpnext.ERPNextReq("GET", itemEndpoint, nil)
	if err != nil {
		fmt.Println("ERROR GET ITEMS ASLI:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data Item dari ERPNext"})
	}

	var rawItems RawItemsResponse
	if err := json.Unmarshal(itemRes, &rawItems); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal parsing data Item"})
	}


	baseURL := os.Getenv("ERPNEXT_URL")
	baseURL = strings.TrimSuffix(baseURL, "/")

	itemGroupMap := make(map[string]*models.ItemGroup)
	templateMap := make(map[string]*models.ItemTemplate)
	var templatesWithVariants []string

	for _, item := range rawItems.Data {
		if strings.HasPrefix(strings.ToUpper(item.Name), "RM-") {
			continue
		}

		if item.VariantOf == "" {
			if itemGroupMap[item.ItemGroup] == nil {
				itemGroupMap[item.ItemGroup] = &models.ItemGroup{
					ItemGroupName: item.ItemGroup,
					Templates:     []*models.ItemTemplate{},
				}
			}


			fullImageURL := ""
			if item.Image != "" {
				// Jaga-jaga jika di ERPNext gambarnya sudah berupa link utuh (http/https)
				if strings.HasPrefix(item.Image, "http") {
					fullImageURL = item.Image
				} else {
					// Pastikan diawali slash (/) sebelum digabung
					if !strings.HasPrefix(item.Image, "/") {
						fullImageURL = baseURL + "/" + item.Image
					} else {
						fullImageURL = baseURL + item.Image
					}
				}
			}

			t := &models.ItemTemplate{
				Name:       item.Name,
				ItemName:   item.ItemName,
				ImageURL:   fullImageURL, // <-- Masukkan URL yang sudah jadi di sini
				Attributes: []models.ItemAttribute{},
			}

			itemGroupMap[item.ItemGroup].Templates = append(itemGroupMap[item.ItemGroup].Templates, t)
			templateMap[item.Name] = t

			if item.HasVariants == 1 {
				templatesWithVariants = append(templatesWithVariants, item.Name)
			}
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var attrCache sync.Map

	for _, tmplName := range templatesWithVariants {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			safeName := url.PathEscape(name)
			detailEndpoint := `/api/resource/Item/` + safeName

			detailRes, err := erpnext.ERPNextReq("GET", detailEndpoint, nil)
			if err != nil {
				return
			}

			var detailData RawItemDetailResponse
			if err := json.Unmarshal(detailRes, &detailData); err == nil {
				var finalAttrs []models.ItemAttribute

				for _, attr := range detailData.Data.Attributes {
					attrName := attr.Attribute
					var attrValues []models.AttributeValue

					if cached, ok := attrCache.Load(attrName); ok {
						attrValues = cached.([]models.AttributeValue)
					} else {
						safeAttrName := url.PathEscape(attrName)
						masterEndpoint := `/api/resource/Item Attribute/` + safeAttrName
						masterRes, err := erpnext.ERPNextReq("GET", masterEndpoint, nil)

						if err == nil {
							var masterData struct {
								Data struct {
									ItemAttributeValues []struct {
										AttributeValue string `json:"attribute_value"`
									} `json:"item_attribute_values"`
								} `json:"data"`
							}

							if json.Unmarshal(masterRes, &masterData) == nil {
								for _, v := range masterData.Data.ItemAttributeValues {
									attrValues = append(attrValues, models.AttributeValue{
										AttributeValue: v.AttributeValue,
									})
								}
							}
						}
						attrCache.Store(attrName, attrValues)
					}

					finalAttrs = append(finalAttrs, models.ItemAttribute{
						Attribute:       attrName,
						AttributeValues: attrValues,
					})
				}

				mu.Lock()
				if tmpl, exists := templateMap[name]; exists {
					tmpl.Attributes = finalAttrs
				}
				mu.Unlock()
			}
		}(tmplName)
	}

	wg.Wait()

	var finalData []models.ItemGroup
	for _, group := range itemGroupMap {
		finalData = append(finalData, *group)
	}

	return c.JSON(fiber.Map{
		"data": finalData,
	})
}


func GetDetailItem(c *fiber.Ctx) error {
	// 1. Tangkap dan decode parameter
	paramItemName, _ := url.QueryUnescape(c.Params("name"))
	if paramItemName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Nama template tidak boleh kosong"})
	}


	searchKeyword := "%" + paramItemName + "%"
	
	tmplFilterArray := []interface{}{
		[]interface{}{"item_name", "like", searchKeyword},
	}
	tmplFilterBytes, _ := json.Marshal(tmplFilterArray)
	
	tmplEndpoint := `/api/resource/Item?filters=` + url.QueryEscape(string(tmplFilterBytes)) + `&fields=["name"]`

	tmplRes, err := erpnext.ERPNextReq("GET", tmplEndpoint, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mencari ID produk di ERPNext"})
	}

	var tmplSearch RawTemplateSearchResponse
	
	if err := json.Unmarshal(tmplRes, &tmplSearch); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal parsing hasil pencarian produk"})
	}

	if len(tmplSearch.Data) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Produk tidak ditemukan di database"})
	}

	actualItemCode := tmplSearch.Data[0].Name


	varFilterArray := []interface{}{
		[]interface{}{"variant_of", "=", actualItemCode},
	}
	varFilterBytes, _ := json.Marshal(varFilterArray)
	
	fieldsParam := `["name","item_name","stock_uom","description"]`
	variantEndpoint := `/api/resource/Item?filters=` + url.QueryEscape(string(varFilterBytes)) + `&fields=` + url.QueryEscape(fieldsParam)

	res, err := erpnext.ERPNextReq("GET", variantEndpoint, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil varian produk"})
	}

	var rawVariants RawVariantResponse
	
	// ---> INI YANG BARU SAJA DIPERBAIKI <---
	if err := json.Unmarshal(res, &rawVariants); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal parsing data varian"})
	}

	if len(rawVariants.Data) == 0 {
		return c.JSON(fiber.Map{"data": []models.ItemVariant{}})
	}

	var itemCodes []string
	for _, v := range rawVariants.Data {
		itemCodes = append(itemCodes, v.Name)
	}

	prFilterArray := []interface{}{
		[]interface{}{"item_code", "in", itemCodes},
	}
	prFilterBytes, _ := json.Marshal(prFilterArray)
	prFields := `["item_code","min_qty","max_qty","price"]`

	prEndpoint := `/api/resource/Pricing Rule?filters=` + url.QueryEscape(string(prFilterBytes)) + `&fields=` + url.QueryEscape(prFields)
	prRes, prErr := erpnext.ERPNextReq("GET", prEndpoint, nil)

	rulesMap := make(map[string][]models.PricingRule)
	
	if prErr == nil {
		var rawRules RawPricingRuleResponse
		if json.Unmarshal(prRes, &rawRules) == nil {
			for _, r := range rawRules.Data {
				rulesMap[r.ItemCode] = append(rulesMap[r.ItemCode], models.PricingRule{
					MinQty: r.MinQty,
					MaxQty: r.MaxQty,
					Price:  r.Price,
				})
			}
		}
	}

	finalVariants := []models.ItemVariant{}

	for _, variant := range rawVariants.Data {
		vrules := rulesMap[variant.Name]
		if vrules == nil {
			vrules = []models.PricingRule{}
		}

		finalVariants = append(finalVariants, models.ItemVariant{
			VariantName:  variant.ItemName,
			ItemCode:     variant.Name,
			UOM:          variant.StockUOM,
			Description:  variant.Description,
			PricingRules: vrules,
		})
	}

	return c.JSON(fiber.Map{
		"data": finalVariants,
	})
}