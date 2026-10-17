package pagination

import (
	"fmt"
	"net/url"
)

// Link 是分页组件中的一个页码。
type Link struct {
	Page   int
	URL    string
	Active bool
}

// View 是模板直接可用的分页视图。
type View struct {
	Page       int
	TotalPages int
	PrevURL    string
	NextURL    string
	Links      []Link
	HasPrev    bool
	HasNext    bool
}

// NewView 根据页码和 URL 生成器构造分页视图。
func NewView(page, totalPages int, url func(int) string) View {
	if page < 1 {
		page = 1
	}
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	const window = 7
	start := page - window/2
	if start < 1 {
		start = 1
	}
	end := start + window - 1
	if end > totalPages {
		end = totalPages
		start = end - window + 1
		if start < 1 {
			start = 1
		}
	}
	links := make([]Link, 0, end-start+1)
	for i := start; i <= end; i++ {
		links = append(links, Link{Page: i, URL: url(i), Active: i == page})
	}
	v := View{Page: page, TotalPages: totalPages, Links: links, HasPrev: page > 1, HasNext: page < totalPages}
	if v.HasPrev {
		v.PrevURL = url(page - 1)
	}
	if v.HasNext {
		v.NextURL = url(page + 1)
	}
	return v
}

// QueryURL 为 query 式分页生成 URL。
func QueryURL(base string, page int, values url.Values) string {
	next := url.Values{}
	for k, v := range values {
		for _, item := range v {
			next.Add(k, item)
		}
	}
	next.Set("page", fmt.Sprint(page))
	return base + "?" + next.Encode()
}
