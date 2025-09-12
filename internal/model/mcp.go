package model

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type McpReg struct {
	Name        string
	Description string
	ToolOptions []mcp.ToolOption
	Fn          server.ToolHandlerFunc
}

type McpSendMessageInput struct {
	Message string `json:"message"`
	AgentId string `json:"agentId"`
	Account string `json:"account"`
}
