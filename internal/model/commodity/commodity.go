package commodity

import (
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

type Commodity struct {
	Name        string
	Type        price.CommodityType
	Code        string
	Harvest     int
	Grandfather string
}

func All() []Commodity {
	var commodities []Commodity
	viper.UnmarshalKey("commodities", &commodities)
	return commodities
}

func FindByName(name string) Commodity {
	c, _ := lo.Find(All(), func(c Commodity) bool { return c.Name == name })
	return c
}
