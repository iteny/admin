package service

import (
	"context"

	"cms/internal/model"
	"cms/internal/pkg/pagination"
	"cms/internal/repository"
)

// HomeData 是前台首页数据。
type HomeData struct {
	Recommended []model.Article
	Latest      []model.Article
	Top         []model.Article
	Hot         []model.Article
	Tags        []model.Tag
	Categories  []*CategoryNode
	Pages       []model.Page
	Total       int64
	Page        int
	PageSize    int
	TotalPages  int
}

// FrontService 组合前台查询。
type FrontService struct {
	articles   *ArticleService
	categories *CategoryService
	tags       *TagService
	pages      *PageService
}

// NewFrontService 创建前台服务。
func NewFrontService(articles *ArticleService, categories *CategoryService, tags *TagService, pages *PageService) *FrontService {
	return &FrontService{articles: articles, categories: categories, tags: tags, pages: pages}
}

// Home 返回首页数据。
func (s *FrontService) Home(ctx context.Context, page, pageSize int) (*HomeData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 9
	}
	recommended, _, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: 1, PageSize: 5}, Status: string(model.ArticleStatusPublished),
		IsRecommend: boolPtr(true),
	})
	if err != nil {
		return nil, err
	}
	latest, total, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: page, PageSize: pageSize}, Status: string(model.ArticleStatusPublished),
	})
	if err != nil {
		return nil, err
	}
	top, _, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: 1, PageSize: 5}, Status: string(model.ArticleStatusPublished),
		IsTop: boolPtr(true),
	})
	if err != nil {
		return nil, err
	}
	hot, _, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: 1, PageSize: 5, Sort: "views"}, Status: string(model.ArticleStatusPublished),
	})
	if err != nil {
		return nil, err
	}
	categories, err := s.categories.Tree(ctx, true)
	if err != nil {
		return nil, err
	}
	pages, err := s.pages.Published(ctx)
	if err != nil {
		return nil, err
	}
	tags, err := s.tags.List(ctx)
	if err != nil {
		return nil, err
	}
	pagesTotal := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pagesTotal < 1 {
		pagesTotal = 1
	}
	return &HomeData{
		Recommended: recommended, Latest: latest, Top: top, Hot: hot,
		Tags: tags, Categories: categories, Pages: pages,
		Total: total, Page: page, PageSize: pageSize, TotalPages: pagesTotal,
	}, nil
}

// CategoryPage 是分类页数据。
type CategoryPage struct {
	Category   model.Category
	Children   []*CategoryNode
	Articles   []model.Article
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// Category 返回分类页数据。
func (s *FrontService) Category(ctx context.Context, slug string, page, pageSize int) (*CategoryPage, error) {
	category, err := s.categories.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	ids, err := s.categories.DescendantIDs(ctx, category.ID)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 9
	}
	list, total, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: page, PageSize: pageSize}, Status: string(model.ArticleStatusPublished),
		CategoryIDs: ids,
	})
	if err != nil {
		return nil, err
	}
	all, err := s.categories.Tree(ctx, true)
	if err != nil {
		return nil, err
	}
	children := findCategoryChildren(all, category.ID)
	pagesTotal := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pagesTotal < 1 {
		pagesTotal = 1
	}
	return &CategoryPage{Category: *category, Children: children, Articles: list, Total: total, Page: page, PageSize: pageSize, TotalPages: pagesTotal}, nil
}

// TagPage 是标签页数据。
type TagPage struct {
	Tag        model.Tag
	Articles   []model.Article
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// Tag 返回标签页数据。
func (s *FrontService) Tag(ctx context.Context, slug string, page, pageSize int) (*TagPage, error) {
	tag, err := s.tags.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 9
	}
	list, total, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: page, PageSize: pageSize}, Status: string(model.ArticleStatusPublished),
		TagID: tag.ID,
	})
	if err != nil {
		return nil, err
	}
	pagesTotal := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pagesTotal < 1 {
		pagesTotal = 1
	}
	return &TagPage{Tag: *tag, Articles: list, Total: total, Page: page, PageSize: pageSize, TotalPages: pagesTotal}, nil
}

// Search 返回搜索结果。
func (s *FrontService) Search(ctx context.Context, keyword string, page, pageSize int) ([]model.Article, int64, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 9
	}
	list, total, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: page, PageSize: pageSize, Keyword: keyword}, Status: string(model.ArticleStatusPublished),
	})
	pages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if pages < 1 {
		pages = 1
	}
	return list, total, pages, err
}

// ArticleDetail 是文章详情数据。
type ArticleDetail struct {
	Article  *model.Article
	Prev     *model.Article
	Next     *model.Article
	Related  []model.Article
	Category *model.Category
}

// Article 返回文章详情，并增加浏览量。
func (s *FrontService) Article(ctx context.Context, slug string) (*ArticleDetail, error) {
	article, err := s.articles.GetPublished(ctx, slug)
	if err != nil {
		return nil, err
	}
	_ = s.articles.View(ctx, article.ID)
	article.ViewCount++
	prev, next, err := s.articles.PrevNext(ctx, article)
	if err != nil {
		return nil, err
	}
	related, err := s.articles.Related(ctx, article.CategoryID, article.ID)
	if err != nil {
		return nil, err
	}
	category := article.Category
	return &ArticleDetail{Article: article, Prev: prev, Next: next, Related: related, Category: &category}, nil
}

// Page 按 slug 返回已发布单页。
func (s *FrontService) Page(ctx context.Context, slug string) (*model.Page, error) {
	return s.pages.GetBySlug(ctx, slug)
}

// Navigation 返回前台导航所需的分类和单页。
func (s *FrontService) Navigation(ctx context.Context) ([]*CategoryNode, []model.Page, error) {
	categories, err := s.categories.Tree(ctx, true)
	if err != nil {
		return nil, nil, err
	}
	pages, err := s.pages.Published(ctx)
	return categories, pages, err
}

// Sitemap 返回 sitemap 所需的全部内容。
func (s *FrontService) Sitemap(ctx context.Context) ([]model.Article, []model.Page, []model.Category, []model.Tag, error) {
	articles, _, err := s.articles.List(ctx, repository.ArticleFilter{
		Query: pagination.Query{Page: 1, PageSize: 1000}, Status: string(model.ArticleStatusPublished),
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}
	pages, err := s.pages.Published(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	categories, err := s.categories.List(ctx, true)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	tags, err := s.tags.List(ctx)
	return articles, pages, categories, tags, err
}

// CategoryFeed 返回分类页的分页路径基地址。
func (s *FrontService) CategoryFeed(category *model.Category, page int) string {
	if page <= 1 {
		return "/category/" + category.Slug
	}
	return "/category/" + category.Slug + "/page/" + itoa(page)
}

// TagFeed 返回标签页的分页路径基地址。
func (s *FrontService) TagFeed(tag *model.Tag, page int) string {
	if page <= 1 {
		return "/tag/" + tag.Slug
	}
	return "/tag/" + tag.Slug + "/page/" + itoa(page)
}

func findCategoryChildren(nodes []*CategoryNode, id uint64) []*CategoryNode {
	for _, node := range nodes {
		if node.Category.ID == id {
			return node.Children
		}
		if children := findCategoryChildren(node.Children, id); len(children) > 0 {
			return children
		}
	}
	return nil
}

func boolPtr(v bool) *bool { return &v }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
