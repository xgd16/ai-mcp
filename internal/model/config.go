package model

type ConfigData struct {
	McpServer *McpServerConfig `json:"mcp-server"`
	DbConfig  *DbConfig        `json:"dbConfig"`
}

type McpServerConfig struct {
	Address string `json:"address"`
}

type DbConfig struct {
	Readonly bool `json:"readonly"`
}
