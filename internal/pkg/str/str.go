// Package str 提供 slug、随机串和文本处理工具。
package str

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mozillazg/go-pinyin"
)

// Slug 根据标题生成 URL slug。中文会转成拼音；无法转换时返回 fallback。
func Slug(title string, enabled bool) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "item"
	}

	var b strings.Builder
	if enabled {
		args := pinyin.NewArgs()
		args.Fallback = func(r rune, _ pinyin.Args) []string {
			return []string{string(r)}
		}
		for _, part := range pinyin.Pinyin(title, args) {
			if len(part) > 0 {
				b.WriteString(part[0])
				b.WriteByte(' ')
			}
		}
	} else {
		b.WriteString(title)
	}

	return NormalizeSlug(b.String())
}

// NormalizeSlug 只保留字母、数字，使用连字符连接。
func NormalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case unicode.IsLetter(r), unicode.IsDigit(r):
			// 非拉丁字母无法安全放进 URL，直接跳过。
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "item"
	}
	return out
}

// Random 返回 n 个十六进制字符的随机串。
func Random(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, (n+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", n)
	}
	return hex.EncodeToString(buf)[:n]
}

// Truncate 按 rune 截断字符串，超出时追加省略号。
func Truncate(s string, n int) string {
	if n <= 0 || utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + "…"
}

// ContainsFold 是大小写不敏感的包含判断（处理 nil 安全）。
func ContainsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
