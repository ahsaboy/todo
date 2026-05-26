package service

import (
	"context"
	"strings"

	"todo/internal/models"
	"todo/internal/repository"
)

type TagService struct {
	repo repository.TagRepository
}

func NewTagService(repo repository.TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) Create(ctx context.Context, userID int64, req models.CreateTagRequest) (*models.UserTag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrTagNameEmpty
	}

	color := models.DefaultTagColor
	if req.Color != nil {
		color = strings.TrimSpace(*req.Color)
		if color == "" {
			color = models.DefaultTagColor
		}
	}
	if !models.IsValidHexColor(color) {
		return nil, ErrInvalidTagColor
	}

	icon := ""
	if req.Icon != nil {
		icon = strings.TrimSpace(*req.Icon)
	}
	if !models.IsValidIcon(icon) {
		return nil, ErrInvalidTagIcon
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	existing, err := s.repo.GetByName(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, repository.ErrTagNameTaken
	}

	tag := &models.UserTag{
		UserID:    userID,
		Name:      name,
		Color:     color,
		Icon:      icon,
		SortOrder: sortOrder,
	}
	return s.repo.Create(ctx, tag)
}

func (s *TagService) List(ctx context.Context, userID int64) ([]models.UserTag, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *TagService) GetByID(ctx context.Context, userID, id int64) (*models.UserTag, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// Update 处理标签字段更新。若 name 变更,会在事务内同步该用户所有任务 tags JSON。
func (s *TagService) Update(ctx context.Context, userID, id int64, req models.UpdateTagRequest) (*models.UserTag, error) {
	// 先取出现状,做必要校验
	current, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	// 校验 color / icon
	if req.Color != nil {
		c := strings.TrimSpace(*req.Color)
		if !models.IsValidHexColor(c) {
			return nil, ErrInvalidTagColor
		}
		req.Color = &c
	}
	if req.Icon != nil {
		ic := strings.TrimSpace(*req.Icon)
		if !models.IsValidIcon(ic) {
			return nil, ErrInvalidTagIcon
		}
		req.Icon = &ic
	}

	// 如果改名,走带同步的事务路径
	if req.Name != nil {
		newName := strings.TrimSpace(*req.Name)
		if newName == "" {
			return nil, ErrTagNameEmpty
		}
		if newName != current.Name {
			updated, renameErr := s.repo.RenameWithTaskSync(ctx, id, userID, newName)
			if renameErr != nil {
				return nil, renameErr
			}
			if updated == nil {
				return nil, nil
			}
			// 改名成功后,再更新其它字段(color/icon/sort_order)
			if req.Color != nil || req.Icon != nil || req.SortOrder != nil {
				return s.repo.Update(ctx, id, userID, nil, req.Color, req.Icon, req.SortOrder)
			}
			return updated, nil
		}
		// 名字未变,走普通 Update
		req.Name = nil
	}

	return s.repo.Update(ctx, id, userID, req.Name, req.Color, req.Icon, req.SortOrder)
}

// Delete 删除标签并从该用户所有任务 tags 中摘除。
// 返回是否删除成功、被影响的任务条数。
func (s *TagService) Delete(ctx context.Context, userID, id int64) (bool, int64, error) {
	return s.repo.DeleteWithTaskSync(ctx, id, userID)
}

// ValidateTagsExist 用于 task service 在 Create/Update 任务时校验 tags 中每个 name 都存在。
// 同时去重 / 去空白。返回清洗后的 tag 名数组(保持原顺序但去重)。
func (s *TagService) ValidateTagsExist(ctx context.Context, userID int64, tags []string) ([]string, error) {
	if len(tags) == 0 {
		return []string{}, nil
	}
	names, err := s.repo.GetNamesSet(ctx, userID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(tags))
	clean := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		if _, ok := names[t]; !ok {
			return nil, ErrUnknownTag
		}
		clean = append(clean, t)
	}
	return clean, nil
}
