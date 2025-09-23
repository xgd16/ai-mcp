package mcp

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/mark3labs/mcp-go/mcp"
)

// RunSafeShellCommand 执行安全受限的终端命令
func (s *sMcpTool) RunSafeShellCommand(ctx context.Context, request mcp.CallToolRequest) (out *mcp.CallToolResult, err error) {
	command := request.GetString("command", "")
	if command == "" {
		err = errors.New("command is required")
		return
	}

	// 超时（秒），默认 10 秒，最大 60 秒
	timeoutSeconds := gconv.Int(request.GetString("timeoutSeconds", "10"))
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	if timeoutSeconds > 60 {
		timeoutSeconds = 60
	}

	// 可选工作目录
	cwd := request.GetString("cwd", "")

	// 风险校验
	if err = validateSafeCommand(command); err != nil {
		out = mcp.NewToolResultText(err.Error())
		err = nil
		return
	}

	// 为了避免交互，使用非交互 shell，并由我们禁用危险操作符
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// 使用 /bin/zsh -lc 或 /bin/sh -lc 均可；macOS 默认有 zsh
	// 我们已在 validateSafeCommand 中禁止了管道与重定向等操作符
	cmd := exec.CommandContext(ctxTimeout, "/bin/zsh", "-lc", command)
	if cwd != "" {
		cmd.Dir = cwd
	}

	stdoutBytes := &strings.Builder{}
	stderrBytes := &strings.Builder{}
	cmd.Stdout = stdoutBytes
	cmd.Stderr = stderrBytes

	start := time.Now()
	runErr := cmd.Run()
	durationMs := time.Since(start).Milliseconds()

	killedByTimeout := ctxTimeout.Err() == context.DeadlineExceeded

	// 退出码
	exitCode := 0
	if runErr != nil {
		// 提取退出码（在大多数情况下）
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// 限制输出大小，防止过大返回
	stdout := trimLong(stdoutBytes.String(), 64*1024)
	stderr := trimLong(stderrBytes.String(), 32*1024)

	result := g.Map{
		"stdout":           stdout,
		"stderr":           stderr,
		"exitCode":         exitCode,
		"durationMs":       durationMs,
		"killedByTimeout":  killedByTimeout,
		"timeoutSeconds":   timeoutSeconds,
		"workingDirectory": cmd.Dir,
		"command":          command,
	}

	// 对于非零退出码，仍返回结果文本，而不是返回错误
	out = mcp.NewToolResultText(gjson.MustEncodeString(result))
	return
}

// validateSafeCommand 黑名单规则与操作符禁用
func validateSafeCommand(command string) error {
	normalized := strings.ToLower(strings.TrimSpace(command))

	// 禁用危险操作符与特性（避免复合执行、重定向、替换等）。放开 | 管道符。
	bannedOperators := []string{
		"||", "&&", ";", ">", ">>", "<", "<<", "`", "$(", "&", "2>", "2>>",
	}
	for _, op := range bannedOperators {
		if strings.Contains(normalized, op) {
			return errors.New("命令包含被禁用的操作符: " + op)
		}
	}

	// 允许使用 |，对每个分段分别做首 token 校验
	segments := strings.Split(normalized, "|")
	if len(segments) > 1 {
		// 最多允许 3 个管道（4 个分段）
		if len(segments)-1 > 3 {
			return errors.New("管道分段过多：最多允许 3 个管道")
		}
		for _, seg := range segments {
			segTrim := strings.TrimSpace(seg)
			if segTrim == "" {
				return errors.New("无效的空管道分段")
			}
			if err := validateFirstToken(segTrim); err != nil {
				return err
			}
		}
	} else {
		if err := validateFirstToken(normalized); err != nil {
			return err
		}
	}

	// 禁用高危命令片段
	bannedFragments := []string{
		"rm -rf", ":(){:|:&};:", "mkfs.", "/dev/", "/etc/passwd",
	}
	for _, frag := range bannedFragments {
		if strings.Contains(normalized, frag) {
			return errors.New("命令包含危险片段: " + frag)
		}
	}

	return nil
}

func validateFirstToken(cmd string) error {
	// 禁用高危命令（匹配首 token）
	bannedCommands := []string{
		"rm", "rmdir", "mkfs", "dd", "chmod", "chown", "mv", "shutdown", "reboot",
		"halt", "poweroff", "init", "service", "systemctl", "mount", "umount", "kill",
		"pkill", "killall", "crontab", "useradd", "userdel", "usermod", "groupadd",
		"groupdel", "visudo", "sudo", "su",
	}

	// 仅检查首 token，避免误杀比如 "echo rm"
	firstToken := firstTokenOf(cmd)
	for _, b := range bannedCommands {
		if firstToken == b {
			return errors.New("命令被禁用: " + b)
		}
	}
	return nil
}

func firstTokenOf(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func trimLong(s string, max int) string {
	if len(s) <= max {
		return s
	}
	// 截断并标记
	suffix := "\n...[truncated]"
	if max > len(suffix) {
		return s[:max-len(suffix)] + suffix
	}
	return s[:max]
}
