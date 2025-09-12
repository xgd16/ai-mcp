package utility

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
)

// ConvertAnyToMarkdownTable 将任意JSON数组转换为Markdown表格
func ConvertAnyToMarkdownTable(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("没有数据可转换")
	}

	// 获取所有字段名（表头）
	var headers []string
	seenHeaders := make(map[string]bool)
	for _, item := range data {
		for key := range item {
			if !seenHeaders[key] {
				seenHeaders[key] = true
				headers = append(headers, key)
			}
		}
	}

	// 计算每列的最大宽度
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// 准备表格数据并更新列宽
	var rows [][]string
	for _, item := range data {
		row := make([]string, len(headers))
		for i, header := range headers {
			value := gconv.String(item[header])
			row[i] = value
			if len(value) > colWidths[i] {
				colWidths[i] = len(value)
			}
		}
		rows = append(rows, row)
	}

	// 构建Markdown表格
	var builder strings.Builder

	// 表头
	builder.WriteString("|")
	for i, header := range headers {
		builder.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], header))
	}
	builder.WriteString("\n")

	// 分隔线
	builder.WriteString("|")
	for i := range headers {
		builder.WriteString(fmt.Sprintf("-%s-|", strings.Repeat("-", colWidths[i])))
	}
	builder.WriteString("\n")

	// 数据行
	for _, row := range rows {
		builder.WriteString("|")
		for i, cell := range row {
			builder.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], cell))
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
