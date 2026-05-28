package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"todo/internal/config"
	"todo/internal/utils"
)

var validLogFilename = regexp.MustCompile(`^backend-\d{4}-\d{2}-\d{2}\.log$`)

type SystemLogHandler struct {
	cfg *config.Config
}

func NewSystemLogHandler(cfg *config.Config) *SystemLogHandler {
	return &SystemLogHandler{cfg: cfg}
}

// ListLogFiles 返回日志目录中所有 backend-*.log 文件列表，按日期倒序。
func (h *SystemLogHandler) ListLogFiles(c *gin.Context) {
	logDir := h.resolveLogDir()
	pattern := filepath.Join(logDir, "backend-*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.log_list", err)
		return
	}

	type logFileInfo struct {
		Filename  string `json:"filename"`
		Date      string `json:"date"`
		SizeBytes int64  `json:"size_bytes"`
	}

	files := make([]logFileInfo, 0, len(matches))
	for _, p := range matches {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		name := filepath.Base(p)
		// 从文件名提取日期部分 "backend-2006-01-02.log" → "2006-01-02"
		date := name[len("backend-") : len(name)-len(".log")]
		files = append(files, logFileInfo{
			Filename:  name,
			Date:      date,
			SizeBytes: info.Size(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Date > files[j].Date
	})

	utils.RespondSuccess(c, files)
}

// GetLogEntries 读取指定日志文件的条目，支持级别过滤和分页。
func (h *SystemLogHandler) GetLogEntries(c *gin.Context) {
	filename := c.Param("filename")
	if !validateLogFile(filename) {
		utils.RespondLocalizedError(c, 400, "system.invalid_log_filename")
		return
	}

	logPath := filepath.Join(h.resolveLogDir(), filepath.Base(filename))
	entries, total, err := readLogEntries(logPath, c.Query("level"))
	if err != nil {
		utils.RespondLocalizedInternalError(c, "system.log_read", err)
		return
	}

	page, limit := h.parsePagination(c)
	start := (page - 1) * limit
	if start > len(entries) {
		start = len(entries)
	}
	end := start + limit
	if end > len(entries) {
		end = len(entries)
	}

	utils.RespondPaginated(c, entries[start:end], page, limit, int64(total))
}

// DownloadLogFile 下载原始日志文件。
func (h *SystemLogHandler) DownloadLogFile(c *gin.Context) {
	filename := c.Param("filename")
	if !validateLogFile(filename) {
		utils.RespondLocalizedError(c, 400, "system.invalid_log_filename")
		return
	}

	logPath := filepath.Join(h.resolveLogDir(), filepath.Base(filename))
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		utils.RespondLocalizedError(c, 404, "system.log_file_not_found")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.File(logPath)
}

// --- helpers ---

func validateLogFile(name string) bool {
	return validLogFilename.MatchString(filepath.Base(name))
}

func (h *SystemLogHandler) resolveLogDir() string {
	return h.cfg.Logging.Path
}

func (h *SystemLogHandler) parsePagination(c *gin.Context) (page, limit int) {
	page = 1
	limit = 50
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	return
}

// readLogEntries 读取日志文件所有行，按 level 过滤后返回。
func readLogEntries(path string, levelFilter string) ([]map[string]any, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	levelFilter = strings.ToLower(strings.TrimSpace(levelFilter))
	var entries []map[string]any

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 最大 1MB 行
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal(line, &entry); err != nil {
			// 跳过无法解析的行
			continue
		}

		if levelFilter != "" {
			if lvl, ok := entry["level"].(string); !ok || strings.ToLower(lvl) != levelFilter {
				continue
			}
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	total := len(entries)
	// 返回时取最新在前（倒序）
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, total, nil
}
