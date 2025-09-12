package mcp

import (
	"ai-mcp/internal/consts"
	"ai-mcp/utility"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/mark3labs/mcp-go/mcp"
)

// ExecSql 执行SQL
func (s *sMcpTool) ExecSql(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	sql := request.GetString("sql", "")
	if sql == "" {
		err = errors.New("sql is required")
		return
	}
	sqlOut, err := g.DB().Query(ctx, sql)
	if err != nil {
		outStr := fmt.Sprintf("数据库执行失败：%s", err.Error())
		consts.Logger.Error(ctx, outStr)
		out = mcp.NewToolResultText(outStr)
		err = nil
		return
	}

	respStr, err := utility.ConvertAnyToMarkdownTable(sqlOut.List())
	if err != nil {
		return
	}
	out = mcp.NewToolResultText(respStr)
	return
}

// GetDatabaseInfo 获取数据库信息
func (s *sMcpTool) GetDatabaseInfo(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	dbname := request.GetString("dbname", "")

	// 获取数据库配置信息
	dbConfig := g.DB(dbname).GetConfig()
	if dbConfig == nil {
		err = errors.New("无法获取数据库配置")
		return
	}

	// 构建数据库信息
	dbInfo := g.Map{
		"databaseType": dbConfig.Type,
		"host":         extractHostFromLink(dbConfig.Link),
		"port":         extractPortFromLink(dbConfig.Link),
		"databaseName": extractDatabaseNameFromLink(dbConfig.Link),
		"username":     extractUsernameFromLink(dbConfig.Link),
		"prefix":       dbConfig.Prefix,
		"createdAt":    dbConfig.CreatedAt,
		"updatedAt":    dbConfig.UpdatedAt,
		"debug":        dbConfig.Debug,
	}

	// 测试连接并获取数据库版本信息
	versionQuery := getVersionQuery(dbConfig.Type)
	if versionQuery != "" {
		sqlOut, queryErr := g.DB().Query(ctx, versionQuery)
		if queryErr == nil && sqlOut != nil && len(sqlOut.List()) > 0 {
			versionInfo := sqlOut.List()[0]
			dbInfo["version"] = versionInfo
		}
	}

	// 获取数据库大小（如果支持）
	sizeQuery := getSizeQuery(dbConfig.Type, dbname)
	if sizeQuery != "" {
		sqlOut, queryErr := g.DB().Query(ctx, sizeQuery)
		if queryErr == nil && sqlOut != nil && len(sqlOut.List()) > 0 {
			sizeInfo := sqlOut.List()[0]
			dbInfo["databaseSize"] = sizeInfo
		}
	}

	out = mcp.NewToolResultText(gjson.MustEncodeString(dbInfo))
	return
}

// 从连接字符串中提取主机地址
func extractHostFromLink(link string) string {
	// 简单的解析，实际项目中可能需要更复杂的解析
	// 格式通常是: user:pass@tcp(host:port)/dbname?params
	if link == "" {
		return ""
	}

	// 查找 tcp( 和 ) 之间的内容
	start := strings.Index(link, "tcp(")
	if start == -1 {
		return ""
	}
	start += 4 // 跳过 "tcp("

	end := strings.Index(link[start:], ")")
	if end == -1 {
		return ""
	}

	hostPort := link[start : start+end]
	// 分离主机和端口
	parts := strings.Split(hostPort, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return hostPort
}

// 从连接字符串中提取端口
func extractPortFromLink(link string) string {
	if link == "" {
		return ""
	}

	start := strings.Index(link, "tcp(")
	if start == -1 {
		return ""
	}
	start += 4

	end := strings.Index(link[start:], ")")
	if end == -1 {
		return ""
	}

	hostPort := link[start : start+end]
	parts := strings.Split(hostPort, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// 从连接字符串中提取数据库名称
func extractDatabaseNameFromLink(link string) string {
	if link == "" {
		return ""
	}

	// 查找最后一个 / 之后的内容
	lastSlash := strings.LastIndex(link, "/")
	if lastSlash == -1 {
		return ""
	}

	dbPart := link[lastSlash+1:]
	// 移除查询参数
	questionMark := strings.Index(dbPart, "?")
	if questionMark != -1 {
		dbPart = dbPart[:questionMark]
	}

	return dbPart
}

// 从连接字符串中提取用户名
func extractUsernameFromLink(link string) string {
	if link == "" {
		return ""
	}

	// 查找 @ 之前的内容
	atIndex := strings.Index(link, "@")
	if atIndex == -1 {
		return ""
	}

	userPass := link[:atIndex]
	// 分离用户名和密码
	colonIndex := strings.Index(userPass, ":")
	if colonIndex != -1 {
		return userPass[:colonIndex]
	}
	return userPass
}

// 根据数据库类型获取版本查询语句
func getVersionQuery(dbType string) string {
	switch dbType {
	case "mysql":
		return "SELECT VERSION() as version"
	case "postgresql", "postgres":
		return "SELECT version() as version"
	case "sqlite":
		return "SELECT sqlite_version() as version"
	case "mssql", "sqlserver":
		return "SELECT @@VERSION as version"
	default:
		return ""
	}
}

// 根据数据库类型获取大小查询语句
func getSizeQuery(dbType, dbname string) string {
	switch dbType {
	case "mysql":
		return fmt.Sprintf("SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)' FROM information_schema.tables WHERE table_schema = '%s'", dbname)
	case "postgresql", "postgres":
		return fmt.Sprintf("SELECT pg_size_pretty(pg_database_size('%s')) as size", dbname)
	case "sqlite":
		return "SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()"
	default:
		return ""
	}
}
