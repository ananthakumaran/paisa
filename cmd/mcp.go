package cmd

import (
	"fmt"
	"os"

	"github.com/ananthakumaran/paisa/internal/mcp"
	"github.com/ananthakumaran/paisa/internal/model"
	"github.com/ananthakumaran/paisa/internal/server"
	"github.com/ananthakumaran/paisa/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start an MCP server over stdio for AI agent integration",
	Long: `Start a Model Context Protocol (MCP) server over stdio.

This allows AI agents (e.g. hermes-agent, Claude Desktop, any MCP-compatible
client) to interact with your Paisa data via the standard MCP protocol.

Example agent config (stdio):
  { "command": "paisa", "args": ["mcp"] }

For Docker / remote deployments, use the HTTP transport instead:
  { "url": "http://localhost:7500/mcp" }
`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := utils.OpenDB()
		if err != nil {
			log.Fatal(err)
		}
		model.AutoMigrate(db)

		if os.Getenv("PAISA_DEBUG") == "true" {
			db = db.Debug()
		}

		// Ensure journal is loaded before serving tool calls
		server.Sync(db, server.SyncRequest{Journal: true})

		s := mcp.BuildMCPServer(db)
		fmt.Fprintln(os.Stderr, "Paisa MCP server started (stdio transport)")
		if err := mcpserver.ServeStdio(s); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
