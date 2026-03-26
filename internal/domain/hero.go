package domain

// Hero represents a Deadlock hero from the assets API.
type Hero struct {
	ID               int    `json:"id"`
	ClassName        string `json:"class_name"`
	Name             string `json:"name"`
	PlayerSelectable bool   `json:"player_selectable"`
	Disabled         bool   `json:"disabled"`
	InDevelopment    bool   `json:"in_development"`
	ImageURL         string `json:"image_url,omitempty"`
}

// HeroImages is the nested images object in the heroes API response.
type HeroImages struct {
	IconHeroCard        string `json:"icon_hero_card"`
	IconHeroCardWebp    string `json:"icon_hero_card_webp"`
	IconImageSmall      string `json:"icon_image_small"`
	IconImageSmallWebp  string `json:"icon_image_small_webp"`
	MinimapImage        string `json:"minimap_image"`
	MinimapImageWebp    string `json:"minimap_image_webp"`
	TopBarVerticalImage string `json:"top_bar_vertical_image"`
}

// HeroRaw is the raw API response for hero, used during JSON unmarshalling.
type HeroRaw struct {
	ID               int        `json:"id"`
	ClassName        string     `json:"class_name"`
	Name             string     `json:"name"`
	PlayerSelectable bool       `json:"player_selectable"`
	Disabled         bool       `json:"disabled"`
	InDevelopment    bool       `json:"in_development"`
	Images           HeroImages `json:"images"`
}

// ToDomain converts a raw hero API response to our domain type.
func (r *HeroRaw) ToDomain() Hero {
	img := r.Images.IconImageSmallWebp
	if img == "" {
		img = r.Images.IconImageSmall
	}
	return Hero{
		ID:               r.ID,
		ClassName:        r.ClassName,
		Name:             r.Name,
		PlayerSelectable: r.PlayerSelectable,
		Disabled:         r.Disabled,
		InDevelopment:    r.InDevelopment,
		ImageURL:         img,
	}
}
