package mcp

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/mark3labs/mcp-go/mcp"
)

// Md5Encode 对输入字符串计算 MD5（十六进制小写）
func (s *sMcpTool) Md5Encode(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	text := request.GetString("text", "")
	if text == "" {
		err = errors.New("text is required")
		return
	}
	sum := md5.Sum([]byte(text))
	out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
		"md5": hex.EncodeToString(sum[:]),
	}))
	return
}

// Base64Encode 对输入字符串进行 Base64 编码
func (s *sMcpTool) Base64Encode(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	text := request.GetString("text", "")
	if text == "" {
		err = errors.New("text is required")
		return
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
		"base64": encoded,
	}))
	return
}

// Base64Decode 对输入 Base64 字符串进行解码
func (s *sMcpTool) Base64Decode(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	data := request.GetString("data", "")
	if data == "" {
		err = errors.New("data is required")
		return
	}
	decodedBytes, decErr := base64.StdEncoding.DecodeString(data)
	if decErr != nil {
		out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
			"error": "invalid base64",
		}))
		err = nil
		return
	}
	out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
		"text": string(decodedBytes),
	}))
	return
}

// JwtParse 解析 JWT（不校验签名），输出 header/payload/rawHeader/rawPayload/rawSignature
func (s *sMcpTool) JwtParse(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	token := request.GetString("token", "")
	if token == "" {
		err = errors.New("token is required")
		return
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
			"error": "invalid jwt format",
		}))
		err = nil
		return
	}

	headerBytes, errH := base64.RawURLEncoding.DecodeString(parts[0])
	payloadBytes, errP := base64.RawURLEncoding.DecodeString(parts[1])
	// 第三段是签名，保持原样
	if errH != nil || errP != nil {
		out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
			"error": "invalid jwt base64 segments",
		}))
		err = nil
		return
	}

	// 解析为 JSON 对象（如果是 JSON）
	headerJson, _ := gjson.DecodeToJson(headerBytes)
	payloadJson, _ := gjson.DecodeToJson(payloadBytes)

	result := g.Map{
		"rawHeader":    parts[0],
		"rawPayload":   parts[1],
		"rawSignature": parts[2],
		"header":       headerJson,  // 可能为 nil
		"payload":      payloadJson, // 可能为 nil
	}

	out = mcp.NewToolResultText(gjson.MustEncodeString(result))
	return
}

// JsonEncode 将任意 JSON 解析为紧凑字符串输出（输入可以是 JSON 字符串）
func (s *sMcpTool) JsonEncode(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	// 允许两种入参形式：raw 任意字符串或 obj JSON 对象（调用方通常传字符串更通用）
	raw := request.GetString("raw", "")
	if raw == "" {
		err = errors.New("raw is required")
		return
	}
	j, jErr := gjson.LoadContent([]byte(raw))
	if jErr != nil {
		out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
			"error": "invalid json",
		}))
		err = nil
		return
	}
	// 紧凑输出；如需 pretty，可扩展参数
	out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
		"json": j.MustToJsonString(),
	}))
	return
}
