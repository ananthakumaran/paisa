package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/accounting"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type ScheduleALSection struct {
	Code    string `json:"code"`
	Section string `json:"section"`
	Details string `json:"details"`
}

var Sections []ScheduleALSection

func init() {
	Sections = []ScheduleALSection{
		{Code: "immovable", Section: "A (1)", Details: "Immovable Assets"},
		{Code: "metal", Section: "B (1) (i)", Details: "Jewellery, bullion etc"},
		{Code: "art", Section: "B (1) (ii)", Details: "Archaeological collections, drawings, painting, sculpture or any work of art"},
		{Code: "vechicle", Section: "B (1) (ii)", Details: "Vehicles, yachts, boats and aircrafts"},
		{Code: "bank", Section: "B (1) (iv) (a)", Details: "Financial assets: Bank (including all deposits)"},
		{Code: "share", Section: "B (1) (iv) (b)", Details: "Financial assets: Shares and securities"},
		{Code: "insurance", Section: "B (1) (iv) (c)", Details: "Financial assets: Insurance policies"},
		{Code: "loan", Section: "B (1) (iv) (d)", Details: "Financial assets: Loans and advances given"},
		{Code: "cash", Section: "B (1) (iv) (e)", Details: "Financial assets: Cash in hand"},
		{Code: "liability", Section: "C (1)", Details: "Liabilities"},
	}
}

type ScheduleALConfig struct {
	Code     string
	Accounts []string
}

type ScheduleALEntry struct {
	Section ScheduleALSection `json:"section"`
	Amount  float64           `json:"amount"`
}

func GetScheduleAL(db *gorm.DB) gin.H {
	var configs []ScheduleALConfig
	viper.UnmarshalKey("schedule_al", &configs)
	now := time.Now()

	postings := query.Init(db).Like("Assets:%").All()
	time := utils.BeginningOfFinancialYear(now)
	postings = lo.Filter(postings, func(p posting.Posting, _ int) bool { return p.Date.Before(time) })

	entries := lo.Map(Sections, func(section ScheduleALSection, _ int) ScheduleALEntry {
		config, found := lo.Find(configs, func(config ScheduleALConfig) bool {
			return config.Code == section.Code
		})

		var amount float64

		if found {
			ps := accounting.FilterByGlob(postings, config.Accounts)
			amount = accounting.CostBalance(ps)
		} else {
			amount = 0
		}

		return ScheduleALEntry{
			Section: section,
			Amount:  amount,
		}
	})
	return gin.H{"schedule_al_entries": entries, "date": time.AddDate(0, 0, -1)}
}
