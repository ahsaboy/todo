package oauth

import (
	"context"

	"todo/internal/config"
)

// Registry 持有所有 OAuth provider 实例。
type Registry struct {
	providers map[string]Provider
}

// NewRegistry 根据配置创建 provider 注册表。LinuxDo 需要初始化 OIDC discovery。
func NewRegistry(cfg *config.Config) *Registry {
	r := &Registry{providers: make(map[string]Provider)}

	gh := NewGitHubProvider(cfg.OAuth.GitHub)
	r.providers["github"] = gh

	gg := NewGoogleProvider(cfg.OAuth.Google)
	r.providers["google"] = gg

	ld := NewLinuxDoProvider(cfg.OAuth.LinuxDo)
	r.providers["linuxdo"] = ld

	return r
}

// Reload 根据新配置重建所有 provider 实例，支持运行时热更新。
func (r *Registry) Reload(ctx context.Context, cfg *config.Config) {
	r.providers["github"] = NewGitHubProvider(cfg.OAuth.GitHub)
	r.providers["google"] = NewGoogleProvider(cfg.OAuth.Google)
	ld := NewLinuxDoProvider(cfg.OAuth.LinuxDo)
	if ld.IsEnabled() {
		ld.Init(ctx)
	}
	r.providers["linuxdo"] = ld
}

// Init 初始化需要预连接的 provider（如 LinuxDo OIDC discovery）。
func (r *Registry) Init(ctx context.Context) error {
	if ld, ok := r.providers["linuxdo"].(*LinuxDoProvider); ok && ld.IsEnabled() {
		if err := ld.Init(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Get 返回指定名称的 provider。
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// EnabledProviders 返回所有已启用的 provider 列表。
func (r *Registry) EnabledProviders() []Provider {
	var result []Provider
	for _, p := range r.providers {
		if p.IsEnabled() {
			result = append(result, p)
		}
	}
	return result
}

// ProviderDisplayInfo 返回 provider 的展示信息。
type ProviderDisplayInfo struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

// GetDisplayInfo 返回前端展示所需的 provider 列表。
func (r *Registry) GetDisplayInfo() []ProviderDisplayInfo {
	labels := map[string]string{
		"github":  "GitHub",
		"google":  "Google",
		"linuxdo": "LinuxDo",
	}
	icons := map[string]string{
		"github":  "github",
		"google":  "chrome",
		"linuxdo": "terminal",
	}

	var result []ProviderDisplayInfo
	for _, p := range r.providers {
		if p.IsEnabled() {
			name := p.Name()
			result = append(result, ProviderDisplayInfo{
				Name:  name,
				Label: labels[name],
				Icon:  icons[name],
			})
		}
	}
	return result
}
