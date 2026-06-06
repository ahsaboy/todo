package service

import (
	"context"
	"encoding/json"
	"fmt"

	"todo/internal/config"
	"todo/internal/repository"
)

type AppConfigService struct {
	repo repository.AppConfigRepository
}

func NewAppConfigService(repo repository.AppConfigRepository) *AppConfigService {
	return &AppConfigService{repo: repo}
}

// Set 校验并持久化单个配置项(只写库,不改运行中的配置;热生效由 handler 单独应用)。
func (s *AppConfigService) Set(ctx context.Context, key string, value any, adminID int64) error {
	spec := config.RegistryByKey()[key]
	if spec == nil {
		return ErrConfigKeyUnknown
	}
	if !spec.Editable {
		return ErrConfigKeyLocked
	}
	if spec.Type == config.TypeEnum {
		sv, ok := value.(string)
		if !ok || !containsString(spec.Enum, sv) {
			return ErrConfigValueInvalid
		}
	}
	// 在临时 Config 上试跑 Set,仅做类型校验,不影响运行中的配置。
	var tmp config.Config
	if err := spec.Set(&tmp, value); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigValueInvalid, err)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConfigValueInvalid, err)
	}
	return s.repo.Upsert(ctx, key, string(raw), adminID)
}

// LoadAll 返回数据库中存储的全部配置覆盖(key→JSON value)。
func (s *AppConfigService) LoadAll(ctx context.Context) (map[string]string, error) {
	return s.repo.LoadAll(ctx)
}

// Reset 删除某个 key 的数据库覆盖(恢复默认,回退配置文件/环境变量)。
func (s *AppConfigService) Reset(ctx context.Context, key string, adminID int64) error {
	spec := config.RegistryByKey()[key]
	if spec == nil {
		return ErrConfigKeyUnknown
	}
	if !spec.Editable {
		return ErrConfigKeyLocked
	}
	return s.repo.Delete(ctx, key)
}

func containsString(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}
