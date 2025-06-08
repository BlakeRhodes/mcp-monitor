// logs/tool.go
package logs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/hpcloud/tail"
)

func NewTool() mcp.Tool {
	return mcp.NewTool("get_log_info",
		mcp.WithDescription("Read recent lines from a log file with optional substring or regex filtering."),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("Path to log file to read (e.g. /var/log/syslog)"),
		),
		mcp.WithNumber("lines",
			mcp.Description("Number of lines from the end (default 20)"),
		),
		mcp.WithString("filter",
			mcp.Description("Case-sensitive substring to include only matching lines."),
		),
		mcp.WithString("regex",
			mcp.Description("Regular expression (Go syntax) to include only lines matching the pattern."),
		),
	)
}

// Helper to read last N lines from journalctl (systemd journal)
func readJournalctlLines(lines int) ([]string, error) {
	cmd := exec.Command("journalctl", "-n", strconv.Itoa(lines))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("journalctl command failed: %v", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	result := []string{}
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning journalctl output: %v", err)
	}
	return result, nil
}

func Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments

	// Required: file (string)
	file, _ := args["file"].(string)
	if file == "" {
		return mcp.NewToolResultError("Parameter 'file' is required."), nil
	}

	// Optional: lines (default 20)
	lines := 20
	if v, ok := args["lines"]; ok {
		switch t := v.(type) {
		case float64:
			lines = int(t)
		case float32:
			lines = int(t)
		case int:
			lines = t
		case int64:
			lines = int(t)
		case string:
			i, err := strconv.Atoi(t)
			if err == nil {
				lines = i
			}
		}
		if lines < 1 {
			lines = 20
		}
	}

	// Optional: substring filter
	filter := ""
	if v, ok := args["filter"]; ok {
		filter, _ = v.(string)
	}

	// Optional: regex filter
	regexStr := ""
	if v, ok := args["regex"]; ok {
		regexStr, _ = v.(string)
	}
	var re *regexp.Regexp
	if regexStr != "" {
		var err error
		re, err = regexp.Compile(regexStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid regex pattern: %v", err)), nil
		}
	}

	var rawLines []string
	// Try file-based logs first
	stat, err := os.Stat(file)
	if err == nil && !stat.IsDir() {
		// Use hpcloud/tail for local file
		tailConf := tail.Config{
			Follow:    false,
			ReOpen:    false,
			MustExist: true,
			Poll:      true,
		}
		tailer, err := tail.TailFile(file, tailConf)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to open log file: %v", err)), nil
		}
		defer tailer.Cleanup()

		for line := range tailer.Lines {
			if line == nil {
				break
			}
			rawLines = append(rawLines, line.Text)
			if len(rawLines) > lines {
				rawLines = rawLines[1:]
			}
		}
	} else {
		// Fallback: if /var/log/syslog (or any file not found) and on systemd system, use journalctl
		if file == "/var/log/syslog" || (os.IsNotExist(err) && strings.Contains(file, "/var/log")) {
			rawLines, err = readJournalctlLines(lines)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to read logs using journalctl: %v", err)), nil
			}
		} else {
			return mcp.NewToolResultError(fmt.Sprintf("File not found or is not a regular file: %s", file)), nil
		}
	}

	// Apply filtering
	buffer := make([]string, 0, lines)
	for _, text := range rawLines {
		if re != nil {
			if !re.MatchString(text) {
				continue
			}
		} else if filter != "" {
			if !strings.Contains(text, filter) {
				continue
			}
		}
		buffer = append(buffer, text)
		if len(buffer) > lines {
			buffer = buffer[1:]
		}
	}
	outputLines := buffer

	summary := map[string]interface{}{
		"file":      file,
		"lines":     outputLines,
		"line_count": len(outputLines),
		"filter":    filter,
		"regex":     regexStr,
	}
	data, _ := json.MarshalIndent(summary, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

