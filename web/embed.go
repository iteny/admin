// Package web 暴露内嵌的模板和静态资源。
package web

import "embed"

// FS 包含 web/template 与 web/static，构建后随二进制分发。
//
//go:embed template static
var FS embed.FS
