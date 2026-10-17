package service

import (
	"context"
	"fmt"
	"strings"

	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/str"
	"cms/internal/repository"
)

// CategoryInput 是分类表单输入。
type CategoryInput struct {
	ParentID       uint64
	Name           string
	Slug           string
	Description    string
	CoverURL       string
	Sort           int
	Status         int8
	Template       string
	SeoTitle       string
	SeoKeywords    string
	SeoDescription string
}

// CategoryNode 是带文章数的分类树节点。
type CategoryNode struct {
	Category     model.Category
	Children     []*CategoryNode
	Depth        int
	ArticleCount int64
}

// CategoryOption 是下拉选项。
type CategoryOption struct {
	ID    uint64
	Name  string
	Depth int
}

// CategoryService 负责分类业务。
type CategoryService struct {
	repo *repository.CategoryRepo
}

// NewCategoryService 创建分类服务。
func NewCategoryService(repo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

// List 返回分类列表。
func (s *CategoryService) List(ctx context.Context, onlyEnabled bool) ([]model.Category, error) {
	return s.repo.List(ctx, onlyEnabled)
}

// Tree 返回分类树和文章数。
func (s *CategoryService) Tree(ctx context.Context, onlyEnabled bool) ([]*CategoryNode, error) {
	list, err := s.repo.List(ctx, onlyEnabled)
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.ArticleCounts(ctx)
	if err != nil {
		return nil, err
	}
	nodes := make(map[uint64]*CategoryNode, len(list))
	for _, item := range list {
		nodes[item.ID] = &CategoryNode{Category: item, ArticleCount: counts[item.ID]}
	}
	var roots []*CategoryNode
	for _, item := range list {
		node := nodes[item.ID]
		if item.ParentID == 0 {
			roots = append(roots, node)
			continue
		}
		if parent, ok := nodes[item.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	var setDepth func(*CategoryNode, int)
	setDepth = func(n *CategoryNode, depth int) {
		n.Depth = depth
		for _, child := range n.Children {
			setDepth(child, depth+1)
		}
	}
	for _, root := range roots {
		setDepth(root, 0)
	}
	return roots, nil
}

// Options 返回带缩进的下拉选项。
func (s *CategoryService) Options(ctx context.Context) ([]CategoryOption, error) {
	roots, err := s.Tree(ctx, false)
	if err != nil {
		return nil, err
	}
	var out []CategoryOption
	var walk func(*CategoryNode)
	walk = func(n *CategoryNode) {
		out = append(out, CategoryOption{ID: n.Category.ID, Name: n.Category.Name, Depth: n.Depth})
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	return out, nil
}

// DescendantIDs 返回分类自身及全部后代 ID。
func (s *CategoryService) DescendantIDs(ctx context.Context, id uint64) ([]uint64, error) {
	list, err := s.repo.List(ctx, false)
	if err != nil {
		return nil, err
	}
	children := make(map[uint64][]uint64)
	for _, item := range list {
		children[item.ParentID] = append(children[item.ParentID], item.ID)
	}
	var out []uint64
	var walk func(uint64)
	walk = func(current uint64) {
		out = append(out, current)
		for _, child := range children[current] {
			walk(child)
		}
	}
	walk(id)
	return out, nil
}

// Get 按 ID 获取分类。
func (s *CategoryService) Get(ctx context.Context, id uint64) (*model.Category, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// GetBySlug 按 slug 获取启用分类。
func (s *CategoryService) GetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	item, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// Create 创建分类。
func (s *CategoryService) Create(ctx context.Context, in CategoryInput) (*model.Category, error) {
	item, err := s.build(ctx, in, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return item, nil
}

// Update 更新分类。
func (s *CategoryService) Update(ctx context.Context, id uint64, in CategoryInput) (*model.Category, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	item, err := s.build(ctx, in, id)
	if err != nil {
		return nil, err
	}
	item.ID = id
	item.CreatedAt = existing.CreatedAt
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return item, nil
}

// Delete 删除分类，存在子分类或文章时拒绝。
func (s *CategoryService) Delete(ctx context.Context, id uint64) error {
	children, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if children > 0 {
		return apperr.NewValidation("id", "请先删除或移动子分类")
	}
	articles, err := s.repo.CountArticles(ctx, id)
	if err != nil {
		return err
	}
	if articles > 0 {
		return apperr.NewValidation("id", "该分类下还有文章，不能删除")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

// Reorder 按 ID 顺序重排。
func (s *CategoryService) Reorder(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return apperr.NewValidation("ids", "排序列表为空")
	}
	return s.repo.Reorder(ctx, ids)
}

func (s *CategoryService) build(ctx context.Context, in CategoryInput, excludeID uint64) (*model.Category, error) {
	name := SanitizeText(in.Name)
	if name == "" {
		return nil, apperr.NewValidation("name", "分类名称不能为空")
	}
	if len([]rune(name)) > 50 {
		return nil, apperr.NewValidation("name", "分类名称不能超过 50 个字符")
	}
	if in.ParentID == excludeID && excludeID > 0 {
		return nil, apperr.NewValidation("parent_id", "不能选择自己作为父分类")
	}
	if in.ParentID > 0 {
		if _, err := s.repo.GetByID(ctx, in.ParentID); err != nil {
			return nil, apperr.NewValidation("parent_id", "父分类不存在")
		}
		if excludeID > 0 {
			ids, err := s.DescendantIDs(ctx, excludeID)
			if err != nil {
				return nil, err
			}
			for _, id := range ids {
				if id == in.ParentID {
					return nil, apperr.NewValidation("parent_id", "不能把分类移动到自己的子分类下")
				}
			}
		}
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = str.Slug(name, true)
	} else {
		slug = str.NormalizeSlug(slug)
	}
	for i := 0; i < 100; i++ {
		candidate := slug
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", slug, i+1)
		}
		exists, err := s.repo.SlugExists(ctx, candidate, excludeID)
		if err != nil {
			return nil, err
		}
		if !exists {
			slug = candidate
			break
		}
	}
	status := in.Status
	if status != model.CategoryStatusDisabled {
		status = model.CategoryStatusEnabled
	}
	return &model.Category{
		ParentID:       in.ParentID,
		Name:           name,
		Slug:           slug,
		Description:    SanitizeText(in.Description),
		CoverURL:       strings.TrimSpace(in.CoverURL),
		Sort:           in.Sort,
		Status:         status,
		Template:       SanitizeText(in.Template),
		SeoTitle:       SanitizeText(in.SeoTitle),
		SeoKeywords:    SanitizeText(in.SeoKeywords),
		SeoDescription: SanitizeText(in.SeoDescription),
	}, nil
}
