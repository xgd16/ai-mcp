# ai-mcp

一个用 Go 构建的 MCP（Model Context Protocol）服务端，基于 `mark3labs/mcp-go` 与 GoFrame v2。项目内置多种可即用的工具（Tools），可通过 MCP 客户端（如 Cursor、继续集成的 IDE/Agent）进行调用。

## 功能特性
- **SSE 服务**：提供基于 Server-Sent Events 的 MCP 服务端；
- **工具集合**：开箱即用的多种工具（见下方“内置工具”）；
- **日志输出**：支持文件与控制台日志，格式与级别可配；
- **数据库支持**：通过 GoFrame gdb，支持 MySQL 等常见数据库；
- **简洁配置**：使用 `config.yaml` 管理服务、日志、数据库等配置。

## 目录结构
```
/Users/xuguodong/code/go/ai-mcp
├── main.go
├── config.yaml
├── makefile
├── internal/
│   ├── consts/
│   │   └── config.go
│   ├── mcp/
│   │   ├── handler.go
│   │   ├── mcp_tool_shell.go
│   │   ├── mcp_tool_db.go
│   │   └── mcp_tool_time.go
│   └── model/
│       ├── config.go
│       └── mcp.go
├── utility/
│   ├── db_data_conv.go
│   └── err.go
└── logs/
```

## 环境要求
- Go `1.25.1`（见 `go.mod`）
- 已安装 `make`（可选）
- 如需数据库相关工具，建议可访问的 MySQL 实例

## 快速开始
1) 获取依赖
```bash
go mod download
```

2) 构建可执行文件
- 使用 `make`：
```bash
make build
```
- 或直接使用 go（开启 greenteagc 实验特性与瘦身）：
```bash
GOEXPERIMENT=greenteagc go build -ldflags="-s -w" -o mcp-server
```

3) 配置
编辑 `config.yaml`：
```yaml
mcp-server:
  address: "127.0.0.1:18232"

logger:
  path: "./logs"
  file: "{Y-m-d}.log"
  prefix: "ai-mcp"
  level: "all"
  timeFormat: "2006-01-02 15:04:05"
  ctxKeys: []
  header: true
  stdout: true

database:
  default:
    type: "mysql"
    link: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb3&parseTime=true&loc=Local"
    prefix: ""
    createdAt: "createTime"
    updatedAt: "updateTime"
    debug: true
```
> 提示：仓库默认配置包含示例数据库连接，请替换为你自己的连接信息，避免敏感信息泄露。

4) 运行
```bash
./mcp-server
```
启动后，终端会打印：
```
MCP SSE服务已启动地址: http://127.0.0.1:18232/sse
```

5) 连接到 MCP 客户端
在支持 MCP 的客户端中新增一个服务端配置，指向上述 SSE 地址（例如 `http://127.0.0.1:18232/sse`）。不同客户端配置方式略有差异，请参考对应客户端文档。

## 内置工具（Tools）
以下工具名称与参数定义自 `internal/mcp/handler.go` 注册，处理函数位于对应 `mcp_tool_*.go` 文件：

- `RunSafeShellCommand`：安全执行终端命令（禁用危险操作符、限制超时与管道数）。
  - 参数：
    - `command`(必填)：命令文本，最多 3 个 `|` 管道，禁止 `&&`/`||`/重定向等；
    - `timeoutSeconds`(可选)：超时秒，默认 10，最大 60；
    - `cwd`(可选)：工作目录。

- `Md5Encode`：对给定文本进行 MD5（小写十六进制）。
  - 参数：`text`(必填)

- `Base64Encode`：文本转 Base64。
  - 参数：`text`(必填)

- `Base64Decode`：Base64 解码为文本。
  - 参数：`data`(必填)

- `JwtParse`：解析 JWT（不校验签名），返回 header 与 payload。
  - 参数：`token`(必填)

- `JsonEncode`：校验并压缩 JSON 字符串。
  - 参数：`raw`(必填)

- `SQL_Actuator`：把自然语言转 SQL 的上层调用者可将 SQL 传入本工具执行，结果以 Markdown 表格返回。
  - 参数：`sql`(必填)

- `GetDatabaseInfo`：返回数据库类型、主机、端口、库名、用户名、版本、大小等信息。
  - 参数：`dbname`(可选)

- `NowTime`：获取当前时间信息与毫秒时间戳。

- `TimestampToDateTime`：将时间戳转换为日期时间字符串。
  - 参数：`timestamp`(必填)

- `GetCalendarDays`：获取指定年与月的所有日期信息（是否周末、英文月名等）。
  - 参数：`year`、`month`(必填)

## 日志
- 日志配置位于 `config.yaml` 的 `logger` 段；
- 当 `path` 配置为目录时，会按 `file` 模板写日志；
- `stdout: true` 时会同时输出到终端。

## 开发说明
- 模块名：`ai-mcp`（见 `go.mod`）
- 主要依赖：
  - `github.com/mark3labs/mcp-go`（MCP 协议实现）
  - `github.com/gogf/gf/v2`（配置、日志、时间、数据库等）
- 入口：`main.go`，启动 SSE 服务，地址来自 `config.yaml` 的 `mcp-server.address`。
- 工具注册：`internal/mcp/handler.go` 遍历 `GetList()` 注册 Tools。

## 常见问题
- 端口被占用：修改 `config.yaml` 中的 `mcp-server.address`。
- 无法连接数据库：确认 `database.default.link` 正确、网络可达，必要时关闭或放通防火墙。
- 客户端连不上：确保使用 SSE 地址（形如 `http://host:port/sse`）。

## 许可协议
MIT License
