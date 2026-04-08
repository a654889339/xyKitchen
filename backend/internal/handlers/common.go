package handlers

import (
	"strconv"
	"strings"

	"xykitchen/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func ctxUser(c *gin.Context) (middleware.CtxUser, bool) {
	return middleware.GetUser(c)
}

func queryInt(c *gin.Context, key string, def int) int {
	s := strings.TrimSpace(c.Query(key))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseID(c *gin.Context, param string) (int, bool) {
	s := c.Param(param)
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// escapeLike 转义 SQL LIKE 中的 % 与 _
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

func firstNonEmptyStr(s, def string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return def
}

func firstNonEmpty(s, def string) string {
	return firstNonEmptyStr(s, def)
}
