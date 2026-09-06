package cmd

import (
	"fmt"
	"os"

	"github.com/ananthakumaran/paisa/internal/mcp"
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	mcpserver "github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the WEB UI",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := utils.OpenDB()
		model.AutoMigrate(db)

		if os.Getenv("PAISA_DEBUG") == "true" {
			db = db.Debug()
		}

		if err != nil {
			log.Fatal(err)
		}

		router := server.Build(db, true)

		// Mount MCP server on /mcp (Streamable HTTP transport).
		// Works identically whether running locally or in Docker — no extra port needed.
		// Protected by the same X-Auth credential check as /api/* when user_accounts are configured.
		mcpH := mcpserver.NewStreamableHTTPServer(mcp.BuildMCPServer(db))
		mcpAuthMW := server.MCPAuthMiddleware(server.NewMCPRateLimiter())
		router.Any("/mcp", mcpAuthMW, gin.WrapH(mcpH))
		router.Any("/mcp/*path", mcpAuthMW, gin.WrapH(mcpH))

		log.Infof("Listening on http://localhost:%d", port)
		log.Infof("MCP server available at http://localhost:%d/mcp", port)
		if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 7500, "port to listen on")
}
