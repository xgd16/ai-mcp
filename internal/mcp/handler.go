package mcp

import (
	"ai-mcp/internal/consts"
	"ai-mcp/internal/model"
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type sMcpTool struct{}

type sMcpHandler struct{}

var (
	McpTool    = &sMcpTool{}
	McpHandler = &sMcpHandler{}
)

func (s *sMcpHandler) GetList() []model.McpReg {
	return []model.McpReg{
		{
			Name:        "RunSafeShellCommand",
			Description: "Execute a terminal command safely with blacklist, operator bans and timeout; supports limited pipes (|)",
			ToolOptions: []mcp.ToolOption{
				mcp.WithString("command",
					mcp.Required(),
					mcp.Description("The terminal command to execute (supports up to 3 pipes, no redirects/logic ops)"),
				),
				mcp.WithString("timeoutSeconds",
					mcp.Description("Timeout seconds (default 10, max 60)"),
				),
				mcp.WithString("cwd",
					mcp.Description("Optional working directory"),
				),
			},
			Fn: McpTool.RunSafeShellCommand,
		},
		{
			Name:        "SQL_Actuator",
			Description: "Convert the user's requirements into SQL statements, execute the SQL statements, and return the execution results",
			ToolOptions: []mcp.ToolOption{
				mcp.WithString("sql",
					mcp.Required(),
					mcp.Description("The SQL statement to be executed"),
				),
			},
			Fn: McpTool.ExecSql,
		},
		{
			Name:        "NowTime",
			Description: "Obtain the current time information，Return the timestamp and date time in the specified time zone",
			ToolOptions: []mcp.ToolOption{
				mcp.WithString("timeZone",
					mcp.Required(),
					mcp.Description("The time zone to be used (e.g. 'Asia/Shanghai')"),
				),
			},
			Fn: McpTool.GetNowTime,
		},
		{
			Name:        "TimestampToDateTime",
			Description: "Convert a timestamp to a date and time",
			ToolOptions: []mcp.ToolOption{
				mcp.WithString("timestamp",
					mcp.Required(),
					mcp.Description("The timestamp to be converted"),
				),
			},
			Fn: McpTool.TimestampToDateTime,
		},
		{
			Name:        "GetCalendarDays",
			Description: "Get all days of a specified year and month",
			ToolOptions: []mcp.ToolOption{
				mcp.WithString("year",
					mcp.Required(),
					mcp.Description("The year (e.g. 2024)"),
				),
				mcp.WithString("month",
					mcp.Required(),
					mcp.Description("The month (1-12)"),
				),
			},
			Fn: McpTool.GetCalendarDays,
		},
		{
			Name:        "GetDatabaseInfo",
			Description: "Get database information including type, name, and connection details",
			ToolOptions: []mcp.ToolOption{
				mcp.WithString("dbname",
					mcp.Description("The database name to query (optional, uses default if not provided)"),
				),
			},
			Fn: McpTool.GetDatabaseInfo,
		},
	}
}

func (s *sMcpHandler) GetMcpFn(item *model.McpReg) (fn server.ToolHandlerFunc) {
	return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
		defer func() {
			if err := recover(); err != nil {
				consts.Logger.Printf(ctx, "panic error %+v", err)
				fn = func(_ context.Context, _ mcp.CallToolRequest) (out *mcp.CallToolResult, outErr error) {
					return mcp.NewToolResultText(fmt.Sprintf("panic error %+v", err)), nil
				}
			}
		}()
		consts.Logger.Printf(ctx, "使用工具 %s 请求内容 %+v", item.Name, request.Params.Arguments)
		if item.Fn == nil {
			return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return mcp.NewToolResultText("处理函数未定义"), nil
			}(ctx, request)
		}
		return item.Fn(ctx, request)
	}
}
