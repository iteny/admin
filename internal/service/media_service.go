package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cms/internal/config"
	"cms/internal/model"
	"cms/internal/pkg/apperr"
	"cms/internal/pkg/hash"
	"cms/internal/pkg/upload"
	"cms/internal/repository"
)

// MediaService 负责上传、去重、缩略图和文件删除。
type MediaService struct {
	repo     *repository.MediaRepo
	base     *config.Config
	settings *SettingService
	root     string
}

// NewMediaService 创建媒体服务。
func NewMediaService(repo *repository.MediaRepo, cfg *config.Config, settings *SettingService) *MediaService {
	return &MediaService{
		repo:     repo,
		base:     cfg,
		settings: settings,
		root:     filepath.Join("storage", "uploads"),
	}
}

func (s *MediaService) uploadConfig() upload.Config {
	maxSize := s.base.App.UploadMaxSize
	if s.settings != nil {
		if raw := s.settings.Get("upload_max_size", ""); raw != "" {
			if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
				maxSize = n
			}
		}
	}
	exts := s.base.App.UploadAllowedExt
	if s.settings != nil {
		if raw := s.settings.Get("upload_allowed_ext", ""); raw != "" {
			exts = strings.Split(raw, ",")
		}
	}
	allowed := make(map[string]bool, len(exts))
	for _, ext := range exts {
		ext = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(ext), "."))
		if ext != "" {
			allowed[ext] = true
		}
	}
	return upload.Config{
		Root:       s.root,
		URLPrefix:  "/uploads",
		MaxSize:    maxSize,
		AllowedExt: allowed,
	}
}

// List 返回媒体列表。
func (s *MediaService) List(ctx context.Context, page, pageSize int, keyword, mimePrefix string) ([]model.Media, int64, error) {
	return s.repo.List(ctx, page, pageSize, keyword, mimePrefix)
}

// Count 返回媒体总数。
func (s *MediaService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

// Get 按 ID 获取媒体。
func (s *MediaService) Get(ctx context.Context, id uint64) (*model.Media, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return item, nil
}

// Upload 保存上传文件；相同内容直接复用已有记录。
func (s *MediaService) Upload(ctx context.Context, file *multipart.FileHeader, uploaderID uint64) (*model.Media, error) {
	if file == nil {
		return nil, apperr.NewValidation("file", "请选择文件")
	}
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()

	uploadCfg := s.uploadConfig()
	data, mimeType, err := uploadCfg.Validate(file.Filename, src, file.Size)
	if err != nil {
		return nil, apperr.NewValidation("file", err.Error())
	}
	if existing, err := s.repo.FindByHash(ctx, hash.SHA256(data)); err == nil {
		return existing, nil
	}
	saved, err := uploadCfg.Save(file.Filename, data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("save upload: %w", err)
	}
	item := &model.Media{
		OriginalName: filepath.Base(file.Filename),
		StoredName:   saved.StoredName,
		Path:         saved.RelPath,
		URL:          saved.URL,
		MimeType:     mimeType,
		Ext:          saved.Ext,
		Size:         saved.Size,
		Width:        saved.Width,
		Height:       saved.Height,
		Hash:         saved.Hash,
		UploaderID:   uploaderID,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		s.removeFiles(saved.RelPath)
		return nil, fmt.Errorf("create media: %w", err)
	}
	return item, nil
}

// Delete 删除记录和磁盘文件。
func (s *MediaService) Delete(ctx context.Context, id uint64) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperr.ErrNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	s.removeFiles(item.Path)
	return nil
}

func (s *MediaService) removeFiles(rel string) {
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return
	}
	full := filepath.Join(s.root, rel)
	_ = os.Remove(full)
	ext := filepath.Ext(full)
	_ = os.Remove(strings.TrimSuffix(full, ext) + "_thumb" + ext)
}
