package domain

// Item represents a purchasable shop item (type="upgrade") from the assets API.
type Item struct {
	ID           int64  `json:"id"`
	ClassName    string `json:"class_name"`
	Name         string `json:"name"`
	ItemSlotType string `json:"item_slot_type"` // "weapon", "vitality", "spirit"
	Tier         int    `json:"item_tier"`       // 1-5
	Cost         int    `json:"cost"`
	IsActive     bool   `json:"is_active_item"`
	Shopable     bool   `json:"shopable"`
	Disabled     bool   `json:"disabled"`
	ImageURL     string `json:"image_url,omitempty"`
}

// ItemRaw is the raw API response for items.
type ItemRaw struct {
	ID           int64  `json:"id"`
	ClassName    string `json:"class_name"`
	Name         string `json:"name"`
	Type         string `json:"type"` // "weapon", "ability", "upgrade"
	ItemSlotType string `json:"item_slot_type"`
	Tier         int    `json:"item_tier"`
	Cost         int    `json:"cost"`
	IsActive     bool   `json:"is_active_item"`
	Shopable     bool   `json:"shopable"`
	Disabled     bool   `json:"disabled"`
	Image        string `json:"image"`
	ImageWebp    string `json:"image_webp"`
	ShopImage    string `json:"shop_image"`
	ShopImageWp  string `json:"shop_image_webp"`
}

// ToDomain converts a raw item API response to our domain type.
// Only "upgrade" type items (shop purchasables) are relevant for WPA.
func (r *ItemRaw) ToDomain() Item {
	img := r.ShopImageWp
	if img == "" {
		img = r.ImageWebp
	}
	if img == "" {
		img = r.Image
	}
	return Item{
		ID:           r.ID,
		ClassName:    r.ClassName,
		Name:         r.Name,
		ItemSlotType: r.ItemSlotType,
		Tier:         r.Tier,
		Cost:         r.Cost,
		IsActive:     r.IsActive,
		Shopable:     r.Shopable,
		Disabled:     r.Disabled,
		ImageURL:     img,
	}
}

// IsShopItem returns true if this is a purchasable upgrade item.
func (r *ItemRaw) IsShopItem() bool {
	return r.Type == "upgrade" && r.Shopable && !r.Disabled
}
