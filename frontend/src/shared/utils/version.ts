/**
 * 获取应用版本
 * 版本通过构建时的 VITE_APP_VERSION 环境变量注入
 */
export function getAppVersion(): string {
  return import.meta.env.VITE_APP_VERSION || 'dev'
}

/**
 * 获取完整的版本信息字符串
 */
export function getVersionString(): string {
  const version = getAppVersion()
  return `v${version}`
}

/**
 * 检查是否为开发版本
 */
export function isDevVersion(): boolean {
  const version = getAppVersion()
  return version === 'dev' || version.includes('dirty')
}