package config

import (
	"fmt"
	"strings"
)

// OrderedPriceCommodities validates price dependencies and orders their sources
// with quote commodities first, using this configuration's default currency.
func (c Config) OrderedPriceCommodities() ([]Commodity, error) {
	byName := make(map[string]Commodity, len(c.Commodities))
	for _, commodity := range c.Commodities {
		if _, exists := byName[commodity.Name]; exists {
			return nil, fmt.Errorf("Duplicate commodity price source: %s", commodity.Name)
		}
		byName[commodity.Name] = commodity
	}

	ordered := make([]Commodity, 0, len(c.Commodities))
	visiting := make(map[string]bool)
	visited := make(map[string]bool)
	var path []string
	var visit func(Commodity) error
	visit = func(commodity Commodity) error {
		if commodity.Name == c.DefaultCurrency || visited[commodity.Name] {
			return nil
		}
		if visiting[commodity.Name] {
			for i, name := range path {
				if name == commodity.Name {
					cycle := append(append([]string{}, path[i:]...), commodity.Name)
					return fmt.Errorf("Circular price dependency: %s. Change the quote commodities so the chain reaches %s.", strings.Join(cycle, " -> "), c.DefaultCurrency)
				}
			}
		}

		visiting[commodity.Name] = true
		path = append(path, commodity.Name)
		quote := commodity.Price.QuoteCommodity(c.DefaultCurrency)
		if quote != c.DefaultCurrency {
			dependency, exists := byName[quote]
			if !exists {
				return fmt.Errorf("%s is quoted in %s, but %s has no configured price source. Add a price source for %s or change %s's quote commodity.", commodity.Name, quote, quote, quote, commodity.Name)
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		path = path[:len(path)-1]
		delete(visiting, commodity.Name)
		visited[commodity.Name] = true
		ordered = append(ordered, commodity)
		return nil
	}

	for _, commodity := range c.Commodities {
		if err := visit(commodity); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
