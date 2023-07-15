package commodity

import (
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/samber/lo"
)

func All() []config.Commodity {
	return config.GetConfig().Commodities
}

func FindByName(name string) config.Commodity {
	c, _ := lo.Find(All(), func(c config.Commodity) bool { return c.Name == name })
	return c
}

func FindByCode(name string) config.Commodity {
	c, _ := lo.Find(All(), func(c config.Commodity) bool { return c.Code == name })
	return c
}

func FindByType(commodityType config.CommodityType) []config.Commodity {
	return lo.Filter(All(), func(c config.Commodity, _ int) bool { return c.Type == commodityType })
}
