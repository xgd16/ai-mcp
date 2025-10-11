package main

import (
	"ai-mcp/internal/consts"
	"fmt"

	sysMcp "ai-mcp/internal/mcp"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	consts.Logger.Info(consts.Ctx, "ai-mcp start")

	// 启动MCP服务
	// Create MCP server
	s := server.NewMCPServer(
		"MCP Server 🚀",
		"1.0.0",
	)

	// Add tool
	fmt.Printf("–––––––––––––––––––––––––––––––––MCP SERVER–––––––––––––––––––––––––––––––––\n\n")
	for _, item := range sysMcp.McpHandler.GetList() {
		fmt.Printf("添加工具 %s - %s\n", item.Name, item.Description)
		s.AddTool(mcp.NewTool(item.Name,
			append([]mcp.ToolOption{
				mcp.WithDescription(item.Description),
			}, item.ToolOptions...)...,
		), sysMcp.McpHandler.GetMcpFn(&item))
	}
	fmt.Printf("\n––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––––\n")

	fmt.Printf("MCP SSE服务已启动地址: http://%s/sse\n", consts.Config.McpServer.Address)

	// Start the stdio server
	if err := server.NewSSEServer(s).Start(consts.Config.McpServer.Address); err != nil {
		panic(err)
	}
}
