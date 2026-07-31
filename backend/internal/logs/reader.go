package logs

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

// Reader 日志文件读取器
// 读取 slog TextHandler 输出的日志文件，按行解析为 model.Log
type Reader struct {
	filePath string
}

// NewReader 创建日志读取器
func NewReader(filePath string) *Reader {
	return &Reader{filePath: filePath}
}

// List 按条件分页读取日志（最新的在前）
func (r *Reader) List(page, pageSize int, level, keyword string) ([]model.Log, int64, error) {
	file, err := os.Open(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Log{}, 0, nil
		}
		return nil, 0, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	// 第一遍：读取全部行，从后向前收集（最新的在前）
	var allLines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 支持长行
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		allLines = append(allLines, line)
	}

	// 从后向前解析（文件末尾是最新的日志）
	var matched []model.Log
	for i := len(allLines) - 1; i >= 0; i-- {
		entry := parseLine(allLines[i])
		if entry == nil {
			continue
		}
		// 级别过滤
		if level != "" && !strings.EqualFold(entry.Level, level) {
			continue
		}
		// 关键词搜索（搜索日志内容）
		if keyword != "" && !strings.Contains(
			strings.ToLower(entry.Content),
			strings.ToLower(keyword),
		) {
			continue
		}
		matched = append(matched, *entry)
	}

	total := int64(len(matched))

	// 分页截取
	start := (page - 1) * pageSize
	if start >= len(matched) {
		return []model.Log{}, total, nil
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}

	result := make([]model.Log, end-start)
	for i := 0; i < len(result); i++ {
		result[i] = matched[start+i]
		// 用行号生成唯一 ID
		result[i].ID = fmt.Sprintf("log-%d", start+i+1)
		result[i].LogPath = r.filePath
	}
	return result, total, nil
}

// parseLine 解析 slog TextHandler 输出的一行日志
// 格式示例:
//
//	time=2026-07-31T12:34:56.789+08:00 level=INFO msg=服务启动 service=devops-dashboard ...
func parseLine(line string) *model.Log {
	fields := make(map[string]string)
	pos := 0

	// 按空格切分，解析 key=value
	for pos < len(line) {
		// 跳过空格和换行
		for pos < len(line) && (line[pos] == ' ' || line[pos] == '\t') {
			pos++
		}
		if pos >= len(line) {
			break
		}

		eq := findNext(line, '=', pos)
		if eq == -1 {
			break
		}
		key := line[pos:eq]

		// 跳过 = 符号
		valueStart := eq + 1
		if valueStart >= len(line) {
			break
		}

		// 找 value 结束位置（下一个空格或行尾）
		valueEnd := findNext(line, ' ', valueStart)
		if valueEnd == -1 {
			valueEnd = len(line)
		}

		value := line[valueStart:valueEnd]
		fields[key] = value
		pos = valueEnd
	}

	// 提取时间：time=2026-07-31T12:34:56.789+08:00 → 07-31 12:34
	timeStr := fields["time"]
	if len(timeStr) >= 19 {
		timeStr = timeStr[5:19] // 截取 "07-31T12:34:56" 部分
		timeStr = strings.Replace(timeStr, "T", " ", 1)
	}

	levelStr := fields["level"]
	if levelStr == "" {
		levelStr = "INFO" // 默认
	}

	msgStr := fields["msg"]
	if msgStr == "" {
		return nil // 没有 msg 的日志行跳过
	}

	// 从日志里提取 service 和 sourceHost
	svc := fields["service"]
	if svc == "" {
		svc = "system"
	}

	return &model.Log{
		Time:       timeStr,
		Level:      levelStr,
		Service:    svc,
		Content:    msgStr,
		SourceHost: fields["hostname"],
	}
}

func findNext(s string, target byte, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == target {
			return i
		}
	}
	return -1
}
