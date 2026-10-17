package service

import (
	"context"
	"fmt"
	"strings"

	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/repository"
)

// MenuInput 是菜单表单输入。
type MenuInput struct {
	ParentID       uint64
	Name           string
	Icon           string
	Path           string
	PermissionCode string
	Sort           int
	IsVisible      bool
	IsExternal     bool
}

// MenuNode 是菜单树节点。
type MenuNode struct {
	Menu     model.Menu
	Children []*MenuNode
	Depth    int
}

// MenuService 负责后台菜单。
type MenuService struct {
	repo *repository.MenuRepo
}

// NewMenuService 创建菜单服务。
func NewMenuService(repo *repository.MenuRepo) *MenuService { return &MenuService{repo: repo} }

// List 返回全部菜单。
func (s *MenuService) List(ctx context.Context) ([]model.Menu, error) {
	return s.repo.List(ctx)
}

// Tree 返回完整菜单树。
func (s *MenuService) Tree(ctx context.Context) ([]*MenuNode, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(list), nil
}

// ForPermissions 按权限返回可见菜单树。
func (s *MenuService) ForPermissions(ctx context.Context, perms map[string]bool) ([]*MenuNode, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.Menu, 0, len(list))
	for _, item := range list {
		if !item.IsVisible {
			continue
		}
		if item.PermissionCode != "" && !perms["*"] && !perms[item.PermissionCode] {
			continue
		}
		filtered = append(filtered, item)
	}
	return buildMenuTree(filtered), nil
}

// Get 按 ID 获取菜单。
func (s *MenuService) Get(ctx context.Context, id uint64) (*model.Menu, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// Create 创建菜单。
func (s *MenuService) Create(ctx context.Context, in MenuInput) (*model.Menu, error) {
	item, err := s.build(ctx, in, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return item, nil
}

// Update 更新菜单。
func (s *MenuService) Update(ctx context.Context, id uint64, in MenuInput) (*model.Menu, error) {
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
		return nil, fmt.Errorf("update menu: %w", err)
	}
	return item, nil
}

// Delete 删除菜单，存在子菜单时拒绝。
func (s *MenuService) Delete(ctx context.Context, id uint64) error {
	n, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperr.NewValidation("id", "请先删除子菜单")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	return nil
}

// Reorder 按给定顺序重排。
func (s *MenuService) Reorder(ctx context.Context, ids []uint64) error {
	if len(ids) == 0 {
		return apperr.NewValidation("ids", "排序列表为空")
	}
	return s.repo.Reorder(ctx, ids)
}

// Options 返回带缩进的菜单下拉选项。
func (s *MenuService) Options(ctx context.Context) ([]*MenuNode, error) {
	roots, err := s.Tree(ctx)
	if err != nil {
		return nil, err
	}
	var out []*MenuNode
	var walk func(*MenuNode)
	walk = func(n *MenuNode) {
		out = append(out, n)
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	return out, nil
}

func (s *MenuService) build(ctx context.Context, in MenuInput, excludeID uint64) (*model.Menu, error) {
	name := SanitizeText(in.Name)
	if name == "" {
		return nil, apperr.NewValidation("name", "菜单名称不能为空")
	}
	if in.ParentID == excludeID && excludeID > 0 {
		return nil, apperr.NewValidation("parent_id", "不能选择自己作为父菜单")
	}
	if in.ParentID > 0 {
		if _, err := s.repo.GetByID(ctx, in.ParentID); err != nil {
			return nil, apperr.NewValidation("parent_id", "父菜单不存在")
		}
		if excludeID > 0 {
			descendants, err := s.descendants(ctx, excludeID)
			if err != nil {
				return nil, err
			}
			for _, id := range descendants {
				if id == in.ParentID {
					return nil, apperr.NewValidation("parent_id", "不能把菜单移动到自己的子菜单下")
				}
			}
		}
	}
	return &model.Menu{
		ParentID:       in.ParentID,
		Name:           name,
		Icon:           strings.TrimSpace(in.Icon),
		Path:           strings.TrimSpace(in.Path),
		PermissionCode: strings.TrimSpace(in.PermissionCode),
		Sort:           in.Sort,
		IsVisible:      in.IsVisible,
		IsExternal:     in.IsExternal,
	}, nil
}

func (s *MenuService) descendants(ctx context.Context, id uint64) ([]uint64, error) {
	list, err := s.repo.List(ctx)
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

func buildMenuTree(list []model.Menu) []*MenuNode {
	nodes := make(map[uint64]*MenuNode, len(list))
	for _, item := range list {
		nodes[item.ID] = &MenuNode{Menu: item}
	}
	var roots []*MenuNode
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
	var setDepth func(*MenuNode, int)
	setDepth = func(n *MenuNode, depth int) {
		n.Depth = depth
		for _, child := range n.Children {
			setDepth(child, depth+1)
		}
	}
	for _, root := range roots {
		setDepth(root, 0)
	}
	return roots
}
