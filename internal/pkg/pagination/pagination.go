// Package pagination 提供列表查询与分页结果。
package pagination

import "math"

// Query 是列表页通用查询参数。
type Query struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Keyword  string `form:"keyword" json:"keyword"`
	Sort     string `form:"sort" json:"sort"`
}

// Normalize 修正分页边界。
func (q *Query) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 15
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
}

// Result 是分页结果。
type Result[T any] struct {
	List       []T   `json:"list"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// New 构造分页结果。
func New[T any](list []T, total int64, page, pageSize int) Result[T] {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 15
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages < 1 {
		pages = 1
	}
	if list == nil {
		list = make([]T, 0)
	}
	return Result[T]{List: list, Total: total, Page: page, PageSize: pageSize, TotalPages: pages}
}

// HasPrev 报告是否有上一页。
func (r Result[T]) HasPrev() bool { return r.Page > 1 }

// HasNext 报告是否有下一页。
func (r Result[T]) HasNext() bool { return r.Page < r.TotalPages }

// Prev 返回上一页页码。
func (r Result[T]) Prev() int {
	if r.HasPrev() {
		return r.Page - 1
	}
	return 1
}

// Next 返回下一页页码。
func (r Result[T]) Next() int {
	if r.HasNext() {
		return r.Page + 1
	}
	return r.TotalPages
}

// Pages 返回用于分页组件的页码窗口（最多 7 个）。
func (r Result[T]) Pages() []int {
	const window = 7
	start := r.Page - window/2
	if start < 1 {
		start = 1
	}
	end := start + window - 1
	if end > r.TotalPages {
		end = r.TotalPages
		start = end - window + 1
		if start < 1 {
			start = 1
		}
	}
	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}
