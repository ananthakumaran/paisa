package model

import (
	"testing"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/stretchr/testify/require"
)

func TestSyncMutualFundPortfolios(t *testing.T) {
	for _, provider := range []string{"in-mfapi", "com-finbodhi-mutualfund"} {
		t.Run(provider, func(t *testing.T) {
			fund := config.Commodity{
				Name: "NIFTY", Type: config.MutualFund,
				Price: config.Price{Provider: provider, Code: "120716"},
			}
			configurePriceSync(t, "INR", fund)
			db := syncTestDB(t)
			fetches := mockPriceRequests(t, map[string]string{
				"": `{"data":[{"name":"Reliance Industries Ltd","isin":"INE002A01018","percentage_to_nav":7.25,"type":"equity"}]}`,
			})

			require.NoError(t, SyncPortfolios(db))
			require.Len(t, fetches.requests, 1)
			holdings := portfolio.GetPortfolios(db, "120716")
			require.Len(t, holdings, 1)
			require.Equal(t, config.MutualFund, holdings[0].CommodityType)
			require.Equal(t, "INE002A01018", holdings[0].SecurityID)
			require.Equal(t, "7.25", holdings[0].Percentage.String())
		})
	}
}
