package cmd

import (
	"github.com/ananthakumaran/paisa/internal/config"
	mutualfund "github.com/ananthakumaran/paisa/internal/model/mutualfund/scheme"
	nps "github.com/ananthakumaran/paisa/internal/model/nps/scheme"
	"github.com/ananthakumaran/paisa/internal/model/portfolio"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/model/template"
	"github.com/ananthakumaran/paisa/internal/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the WEB UI",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := gorm.Open(sqlite.Open(config.GetConfig().DBPath), &gorm.Config{})
		db.AutoMigrate(&nps.Scheme{})
		db.AutoMigrate(&mutualfund.Scheme{})
		db.AutoMigrate(&posting.Posting{})
		db.AutoMigrate(&price.Price{})
		db.AutoMigrate(&portfolio.Portfolio{})
		db.AutoMigrate(&template.Template{})

		if err != nil {
			log.Fatal(err)
		}
		server.Listen(db)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
