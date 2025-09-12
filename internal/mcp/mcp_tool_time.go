package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *sMcpTool) GetNowTime(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	now := gtime.Now()
	// 获取指定时区
	out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
		"datetime":        now.Format("Y-m-d H:i:s"),
		"UnixMillisecond": now.UnixMilli(),
	}))
	return
}

// TimestampToDateTime 时间戳转换为日期时间
func (s *sMcpTool) TimestampToDateTime(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	timestamp := request.GetString("timestamp", "")
	if timestamp == "" {
		err = errors.New("timestamp is empty")
		return
	}
	gtime.NewFromTimeStamp(gconv.Int64(timestamp)).Format("Y-m-d H:i:s")
	out = mcp.NewToolResultText(gjson.MustEncodeString(g.Map{
		"datetime": gtime.NewFromTimeStamp(gconv.Int64(timestamp)).Format("Y-m-d H:i:s"),
	}))
	return
}

// GetCalendarDays 获取指定年月的每一天日期
func (s *sMcpTool) GetCalendarDays(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	year := request.GetString("year", "")
	month := request.GetString("month", "")

	if year == "" {
		err = errors.New("year is required")
		return
	}
	if month == "" {
		err = errors.New("month is required")
		return
	}

	yearInt := gconv.Int(year)
	monthInt := gconv.Int(month)

	// 验证年份和月份的有效性
	if yearInt < 1 || yearInt > 9999 {
		err = errors.New("year must be between 1 and 9999")
		return
	}
	if monthInt < 1 || monthInt > 12 {
		err = errors.New("month must be between 1 and 12")
		return
	}

	// 获取该月的第一天
	firstDay := gtime.NewFromStr(fmt.Sprintf("%d-%02d-01", yearInt, monthInt))
	if firstDay == nil {
		err = errors.New("invalid date")
		return
	}

	// 获取该月的最后一天（下个月的第一天减一天）
	var lastDay *gtime.Time
	if monthInt == 12 {
		lastDay = gtime.NewFromStr(fmt.Sprintf("%d-01-01", yearInt+1)).AddDate(0, 0, -1)
	} else {
		lastDay = gtime.NewFromStr(fmt.Sprintf("%d-%02d-01", yearInt, monthInt+1)).AddDate(0, 0, -1)
	}

	// 生成该月所有日期
	var days []g.Map
	currentDay := firstDay
	for currentDay.Before(lastDay) || currentDay.Equal(lastDay) {
		days = append(days, g.Map{
			"date":        currentDay.Format("Y-m-d"),
			"day":         currentDay.Day(),
			"weekday":     currentDay.Weekday(),
			"weekdayName": currentDay.Format("l"),
			"isWeekend":   currentDay.Weekday() == 0 || currentDay.Weekday() == 6, // 0=Sunday, 6=Saturday
		})
		currentDay = currentDay.AddDate(0, 0, 1)
	}

	result := g.Map{
		"year":      yearInt,
		"month":     monthInt,
		"monthName": firstDay.Format("F"),
		"totalDays": len(days),
		"days":      days,
	}

	out = mcp.NewToolResultText(gjson.MustEncodeString(result))
	return
}
