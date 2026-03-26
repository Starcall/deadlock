package deadlockapi

import (
	"context"
	"fmt"

	"github.com/name/deadlock/internal/domain"
)

// FetchHeroes retrieves all heroes from the assets API.
func (c *Client) FetchHeroes(ctx context.Context) ([]domain.Hero, error) {
	url := fmt.Sprintf("%s/v2/heroes", assetsURL)

	var raw []domain.HeroRaw
	if err := c.getJSON(ctx, url, &raw); err != nil {
		return nil, fmt.Errorf("fetching heroes: %w", err)
	}

	heroes := make([]domain.Hero, 0, len(raw))
	for _, r := range raw {
		if r.Disabled || r.InDevelopment {
			continue
		}
		heroes = append(heroes, r.ToDomain())
	}
	return heroes, nil
}

// FetchAllHeroes retrieves all heroes including disabled/in-development ones.
func (c *Client) FetchAllHeroes(ctx context.Context) ([]domain.Hero, error) {
	url := fmt.Sprintf("%s/v2/heroes", assetsURL)

	var raw []domain.HeroRaw
	if err := c.getJSON(ctx, url, &raw); err != nil {
		return nil, fmt.Errorf("fetching heroes: %w", err)
	}

	heroes := make([]domain.Hero, 0, len(raw))
	for _, r := range raw {
		heroes = append(heroes, r.ToDomain())
	}
	return heroes, nil
}

// FetchItems retrieves all shop items (type=upgrade, shopable) from the assets API.
func (c *Client) FetchItems(ctx context.Context) ([]domain.Item, error) {
	url := fmt.Sprintf("%s/v2/items", assetsURL)

	var raw []domain.ItemRaw
	if err := c.getJSON(ctx, url, &raw); err != nil {
		return nil, fmt.Errorf("fetching items: %w", err)
	}

	items := make([]domain.Item, 0)
	for _, r := range raw {
		if r.IsShopItem() {
			items = append(items, r.ToDomain())
		}
	}
	return items, nil
}

// FetchAllItems retrieves all items from the assets API.
func (c *Client) FetchAllItems(ctx context.Context) ([]domain.ItemRaw, error) {
	url := fmt.Sprintf("%s/v2/items", assetsURL)

	var raw []domain.ItemRaw
	if err := c.getJSON(ctx, url, &raw); err != nil {
		return nil, fmt.Errorf("fetching items: %w", err)
	}
	return raw, nil
}
