package model

type ConfigData struct {
	McpServer *McpServerConfig `json:"mcp-server"`
}

type McpServerConfig struct {
	Address string `json:"address"`
}
