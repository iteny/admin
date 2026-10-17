// Package render 实现基于 html/template 的 Gin 渲染器。
package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	ginrender "github.com/gin-gonic/gin/render"
)

// Renderer 按页面预编译模板集合。
type Renderer struct {
	mu    sync.RWMutex
	sets  map[string]*template.Template
	debug bool
	fsys  fs.FS
	funcs template.FuncMap
}

// New 创建渲染器。fsys 的根目录应为 templates 目录。
func New(fsys fs.FS, debug bool, funcs template.FuncMap) (*Renderer, error) {
	r := &Renderer{debug: debug, fsys: fsys, funcs: funcs}
	if err := r.parseAll(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Renderer) parseAll() error {
	shared, err := r.sharedFiles()
	if err != nil {
		return err
	}
	pages, err := r.pageFiles()
	if err != nil {
		return err
	}
	sets := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		layout, err := r.layoutFor(page)
		if err != nil {
			return err
		}
		files := append(append([]string{}, shared...), layout, page)
		tpl, err := template.New(path.Base(page)).Funcs(r.funcs).ParseFS(r.fsys, files...)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", page, err)
		}
		sets[page] = tpl
	}
	r.mu.Lock()
	r.sets = sets
	r.mu.Unlock()
	return nil
}

func (r *Renderer) sharedFiles() ([]string, error) {
	return fs.Glob(r.fsys, "partials/*.html")
}

func (r *Renderer) pageFiles() ([]string, error) {
	var pages []string
	for _, root := range []string{"admin", "web", "errors"} {
		err := fs.WalkDir(r.fsys, root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(p, ".html") {
				pages = append(pages, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func (r *Renderer) layoutFor(page string) (string, error) {
	switch {
	case strings.HasPrefix(page, "admin/auth/"):
		return "layouts/auth.html", nil
	case page == "admin/media/picker.html":
		return "layouts/fragment.html", nil
	case strings.HasPrefix(page, "admin/"):
		return "layouts/admin.html", nil
	case strings.HasPrefix(page, "web/"), strings.HasPrefix(page, "errors/"):
		return "layouts/web.html", nil
	default:
		return "", fmt.Errorf("no layout for template %s", page)
	}
}

// Instance 实现 gin/render.HTMLRender。
func (r *Renderer) Instance(name string, data any) ginrender.Render {
	if r.debug {
		if err := r.parseAll(); err != nil {
			slog.Error("reparse templates", "error", err, "template", name)
		}
	}
	r.mu.RLock()
	tpl, ok := r.sets[name]
	fallback := r.sets["errors/500.html"]
	r.mu.RUnlock()
	if !ok {
		slog.Error("template not found", "name", name)
		tpl = fallback
	}
	if tpl == nil {
		return errorRender{err: fmt.Errorf("template %q not found", name)}
	}
	return htmlRender{tpl: tpl, data: data}
}

type htmlRender struct {
	tpl  *template.Template
	data any
}

func (h htmlRender) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	if err := h.tpl.ExecuteTemplate(w, "layout", h.data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	return nil
}

func (h htmlRender) WriteContentType(w http.ResponseWriter) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}

type errorRender struct{ err error }

func (e errorRender) Render(w http.ResponseWriter) error {
	http.Error(w, "template error", http.StatusInternalServerError)
	return e.err
}

func (e errorRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

// FuncMap 返回模板函数集合。
func FuncMap(loc *time.Location) template.FuncMap {
	if loc == nil {
		loc = time.Local
	}
	return template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec
		"fmtTime": func(value any) string {
			t, ok := asTime(value)
			if !ok || t.IsZero() {
				return "-"
			}
			return t.In(loc).Format("2006-01-02 15:04")
		},
		"fmtDate": func(value any) string {
			t, ok := asTime(value)
			if !ok || t.IsZero() {
				return "-"
			}
			return t.In(loc).Format("2006-01-02")
		},
		"fmtInputTime": func(value any) string {
			t, ok := asTime(value)
			if !ok || t.IsZero() {
				return ""
			}
			return t.In(loc).Format("2006-01-02T15:04")
		},
		"timeAgo": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "刚刚"
			case d < time.Hour:
				return strconv.Itoa(int(d.Minutes())) + " 分钟前"
			case d < 24*time.Hour:
				return strconv.Itoa(int(d.Hours())) + " 小时前"
			case d < 30*24*time.Hour:
				return strconv.Itoa(int(d.Hours()/24)) + " 天前"
			default:
				return t.In(loc).Format("2006-01-02")
			}
		},
		"truncate": func(s string, n int) string {
			if n <= 0 || utf8.RuneCountInString(s) <= n {
				return s
			}
			return string([]rune(s)[:n]) + "…"
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"seq": func(n int) []int {
			out := make([]int, 0, n)
			for i := 1; i <= n; i++ {
				out = append(out, i)
			}
			return out
		},
		"hasPerm": func(perms map[string]bool, code string) bool {
			if code == "" || perms["*"] {
				return true
			}
			return perms[code]
		},
		"menuActive": func(current, path string) bool {
			if path == "" {
				return false
			}
			if path == "/admin" {
				return current == "/admin"
			}
			return strings.HasPrefix(current, path)
		},
		"hasPrefix": strings.HasPrefix,
		"initial": func(s string) string {
			r := []rune(strings.TrimSpace(s))
			if len(r) == 0 {
				return "U"
			}
			return string(r[0])
		},
		"queryWith": func(values url.Values, key, value string) string {
			next := url.Values{}
			for k, v := range values {
				for _, item := range v {
					next.Add(k, item)
				}
			}
			if value == "" {
				next.Del(key)
			} else {
				next.Set(key, value)
			}
			return next.Encode()
		},
		"statusText": func(status string) string {
			switch status {
			case "published":
				return "已发布"
			case "draft":
				return "草稿"
			case "archived":
				return "已归档"
			default:
				return status
			}
		},
		"statusClass": func(status string) string {
			switch status {
			case "published":
				return "bg-green-lt"
			case "draft":
				return "bg-yellow-lt"
			case "archived":
				return "bg-secondary-lt"
			default:
				return "bg-blue-lt"
			}
		},
		"dict": func(values ...any) map[string]any {
			m := make(map[string]any, len(values)/2)
			for i := 0; i+1 < len(values); i += 2 {
				m[fmt.Sprint(values[i])] = values[i+1]
			}
			return m
		},
		"default": func(def, value any) any {
			if value == nil || value == "" || value == 0 {
				return def
			}
			return value
		},
		"join": strings.Join,
		"highlight": func(text, keyword string) template.HTML {
			escaped := template.HTMLEscapeString(text)
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				return template.HTML(escaped) //nolint:gosec
			}
			re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(template.HTMLEscapeString(keyword)))
			return template.HTML(re.ReplaceAllString(escaped, "<mark>$0</mark>")) //nolint:gosec
		},
		"json": func(v any) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(b) //nolint:gosec
		},
		"thumbURL": func(u string) string {
			if u == "" {
				return ""
			}
			ext := path.Ext(u)
			return strings.TrimSuffix(u, ext) + "_thumb" + ext
		},
	}
}

func asTime(value any) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case *time.Time:
		if v == nil {
			return time.Time{}, false
		}
		return *v, true
	default:
		return time.Time{}, false
	}
}

var _ ginrender.HTMLRender = (*Renderer)(nil)
