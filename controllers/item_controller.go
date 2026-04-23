package controllers

import (
	"bikincetak-api/models"
	"bikincetak-api/erpnext"
	"encoding/json"
	"net/url"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type RawItemsResponse struct {
	Data []struct {
		Name        string `json:"name"`
		ItemName    string `json:"item_name"`
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

func GetItems(c *fiber.Ctx) error {
	itemEndpoint := `/api/resource/Item?fields=["name","item_name","has_variants","variant_of"]&limit=1000`
	itemRes, err := erpnext.ERPNextReq("GET", itemEndpoint, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data Item dari ERPNext"})
	}

	var rawItems RawItemsResponse
	if err := json.Unmarshal(itemRes, &rawItems); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal parsing data Item"})
	}

	templateMap := make(map[string]*models.Template)
	var templatesWithVariants []string

	for _, item := range rawItems.Data {
		if strings.HasPrefix(strings.ToUpper(item.Name), "RM-") {
			continue
		}

		if item.VariantOf == "" {
			t := models.Template{
				Name:        item.Name,
				ItemName:    item.ItemName,
				HasVariants: item.HasVariants,
				Attributes:  []models.ItemAttribute{}, 
			}
			templateMap[item.Name] = &t

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
								// Pindahkan nilainya ke struct model kita
								for _, v := range masterData.Data.ItemAttributeValues {
									attrValues = append(attrValues, models.AttributeValue{
										Value: v.AttributeValue,
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
				templateMap[name].Attributes = finalAttrs
				mu.Unlock()
			}
		}(tmplName)
	}

	wg.Wait()

	var finalTemplates []models.Template
	for _, t := range templateMap {
		finalTemplates = append(finalTemplates, *t)
	}

	return c.JSON(fiber.Map{
		"message": "Katalog berhasil dimuat, difilter, dan nilai atribut berhasil diambil",
		"data":    finalTemplates,
	})
}