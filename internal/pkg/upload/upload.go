// Package upload 负责文件类型校验、图片缩放、缩略图生成和磁盘落盘。
package upload

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"

	"cms/internal/pkg/hash"
	"cms/internal/pkg/str"
)

// Saved 描述一次成功保存。
type Saved struct {
	OriginalName string
	StoredName   string
	RelPath      string
	URL          string
	ThumbURL     string
	MimeType     string
	Ext          string
	Size         int64
	Width        int
	Height       int
	Hash         string
}

// Config 是上传配置。
type Config struct {
	Root       string
	URLPrefix  string
	MaxSize    int64
	AllowedExt map[string]bool
}

// Validate 读取并校验文件内容，返回原始字节。
func (c Config) Validate(filename string, r io.Reader, size int64) ([]byte, string, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		return nil, "", fmt.Errorf("文件缺少扩展名")
	}
	if !c.AllowedExt[ext] {
		return nil, "", fmt.Errorf("不支持的文件类型 .%s", ext)
	}
	if size > 0 && size > c.MaxSize {
		return nil, "", fmt.Errorf("文件超过大小限制（最大 %d MB）", c.MaxSize/(1<<20))
	}
	limit := c.MaxSize
	if limit <= 0 {
		limit = 10 << 20
	}
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, "", fmt.Errorf("读取上传文件: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, "", fmt.Errorf("文件超过大小限制（最大 %d MB）", limit/(1<<20))
	}
	mime := http.DetectContentType(data)
	if err := checkMagic(ext, data, mime); err != nil {
		return nil, "", err
	}
	return data, mime, nil
}

func checkMagic(ext string, data []byte, mime string) error {
	switch ext {
	case "jpg", "jpeg":
		if !bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}) {
			return fmt.Errorf("文件内容不是有效的 JPEG")
		}
	case "png":
		if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) {
			return fmt.Errorf("文件内容不是有效的 PNG")
		}
	case "gif":
		if !bytes.HasPrefix(data, []byte("GIF8")) {
			return fmt.Errorf("文件内容不是有效的 GIF")
		}
	case "webp":
		if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
			return fmt.Errorf("文件内容不是有效的 WebP")
		}
	case "svg":
		if !bytes.Contains(data[:min(len(data), 512)], []byte("<svg")) && !bytes.Contains(data[:min(len(data), 512)], []byte("<?xml")) {
			return fmt.Errorf("文件内容不是有效的 SVG")
		}
		if !strings.Contains(mime, "svg") && !strings.Contains(mime, "xml") {
			return fmt.Errorf("文件内容不是有效的 SVG")
		}
	default:
		if !strings.HasPrefix(mime, "image/") {
			return fmt.Errorf("不允许的文件类型")
		}
	}
	return nil
}

// Save 保存文件，按需缩放到最大宽度 1920 并生成 400px 缩略图。
func (c Config) Save(originalName string, data []byte, mime string) (Saved, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(originalName), "."))
	now := time.Now()
	relDir := now.Format("2006/01")
	dir := filepath.Join(c.Root, filepath.FromSlash(relDir))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return Saved{}, fmt.Errorf("创建上传目录: %w", err)
	}

	stored := str.Random(12) + "." + ext
	rel := path.Join(relDir, stored)
	filePath := filepath.Join(c.Root, filepath.FromSlash(rel))
	width, height := 0, 0
	thumbURL := ""

	if isResizable(ext) {
		if img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true)); err == nil {
			width, height = img.Bounds().Dx(), img.Bounds().Dy()
			if width > 1920 {
				img = imaging.Resize(img, 1920, 0, imaging.Lanczos)
				width, height = img.Bounds().Dx(), img.Bounds().Dy()
				var buf bytes.Buffer
				if err := imaging.Encode(&buf, img, encodingFor(ext)); err == nil {
					data = buf.Bytes()
				}
			}
			thumb := imaging.Resize(img, 400, 0, imaging.Lanczos)
			thumbName := strings.TrimSuffix(stored, "."+ext) + "_thumb." + ext
			thumbPath := filepath.Join(dir, thumbName)
			if err := imaging.Save(thumb, thumbPath, imaging.JPEGQuality(82)); err == nil {
				thumbURL = strings.TrimRight(c.URLPrefix, "/") + "/" + path.Join(relDir, thumbName)
			}
		}
	}

	if err := os.WriteFile(filePath, data, 0o640); err != nil {
		return Saved{}, fmt.Errorf("写入上传文件: %w", err)
	}
	url := strings.TrimRight(c.URLPrefix, "/") + "/" + rel
	return Saved{
		OriginalName: originalName,
		StoredName:   stored,
		RelPath:      rel,
		URL:          url,
		ThumbURL:     thumbURL,
		MimeType:     mime,
		Ext:          ext,
		Size:         int64(len(data)),
		Width:        width,
		Height:       height,
		Hash:         hash.SHA256(data),
	}, nil
}

func isResizable(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "png", "gif":
		return true
	default:
		return false
	}
}

func encodingFor(ext string) imaging.Format {
	switch ext {
	case "png":
		return imaging.PNG
	case "gif":
		return imaging.GIF
	default:
		return imaging.JPEG
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
