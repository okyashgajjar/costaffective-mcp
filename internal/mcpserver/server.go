package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
)

func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"CostAffective Code Intelligence",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	RegisterTools(s)
	return s
}
