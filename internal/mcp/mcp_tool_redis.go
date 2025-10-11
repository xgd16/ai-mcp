package mcp

import (
	"ai-mcp/internal/consts"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/mark3labs/mcp-go/mcp"
)

// ExecRedisCommand 执行Redis命令
func (s *sMcpTool) ExecRedisCommand(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	// 获取命令和参数
	command := request.GetString("command", "")
	if command == "" {
		err = errors.New("command is required")
		return
	}

	// 获取命令参数
	argsStr := request.GetString("args", "")
	var args []interface{}
	if argsStr != "" {
		var argsArray []string
		if err = gjson.Unmarshal([]byte(argsStr), &argsArray); err != nil {
			errMsg := fmt.Sprintf("参数解析失败: %s", err.Error())
			consts.Logger.Error(ctx, errMsg)
			out = mcp.NewToolResultText(errMsg)
			err = nil
			return
		}
		// 转换为 interface{} 切片
		args = make([]interface{}, len(argsArray))
		for i, v := range argsArray {
			args[i] = v
		}
	}

	// 获取 Redis 连接
	conn, err := g.Redis().Conn(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("Redis连接失败: %s", err.Error())
		consts.Logger.Error(ctx, errMsg)
		out = mcp.NewToolResultText(errMsg)
		err = nil
		return
	}
	defer conn.Close(ctx)

	// 执行 Redis 命令
	result, err := conn.Do(ctx, command, args...)
	if err != nil {
		errMsg := fmt.Sprintf("Redis命令执行失败: %s", err.Error())
		consts.Logger.Error(ctx, errMsg)
		out = mcp.NewToolResultText(errMsg)
		err = nil
		return
	}

	// 格式化返回结果
	resultStr := formatRedisResult(command, result)
	out = mcp.NewToolResultText(resultStr)
	return
}

// formatRedisResult 格式化Redis命令的返回结果
func formatRedisResult(command string, result interface{}) string {
	if result == nil {
		return "(nil)"
	}

	cmd := strings.ToUpper(command)

	switch v := result.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int64:
		return fmt.Sprintf("%d", v)
	case []interface{}:
		// 对于列表类型的返回值
		if len(v) == 0 {
			return "(empty list or set)"
		}

		var builder strings.Builder
		// 特殊处理 HGETALL 等返回 key-value 对的命令
		if cmd == "HGETALL" || cmd == "HKEYS" || cmd == "HVALS" {
			if cmd == "HGETALL" && len(v)%2 == 0 {
				// HGETALL 返回的是 key-value 对
				builder.WriteString("Hash fields:\n")
				for i := 0; i < len(v); i += 2 {
					key := fmt.Sprintf("%v", v[i])
					val := fmt.Sprintf("%v", v[i+1])
					builder.WriteString(fmt.Sprintf("%s: %s\n", key, val))
				}
				return strings.TrimRight(builder.String(), "\n")
			}
		}

		// 普通列表
		builder.WriteString(fmt.Sprintf("Total: %d items\n", len(v)))
		for i, item := range v {
			builder.WriteString(fmt.Sprintf("%d) %v\n", i+1, item))
		}
		return strings.TrimRight(builder.String(), "\n")
	case map[string]interface{}:
		// 对于 map 类型的返回值
		if len(v) == 0 {
			return "(empty map)"
		}
		var builder strings.Builder
		builder.WriteString("Map data:\n")
		for key, val := range v {
			builder.WriteString(fmt.Sprintf("%s: %v\n", key, val))
		}
		return strings.TrimRight(builder.String(), "\n")
	default:
		// 其他类型，使用 gjson 格式化
		jsonStr, err := gjson.EncodeString(result)
		if err != nil {
			return fmt.Sprintf("%v", result)
		}
		return jsonStr
	}
}
