/**
 * 获取当前页面的基础重定向 URI（含路径前缀）。
 * 用于 OAuth 回调，确保反向代理场景下路径前缀不丢失。
 *
 * 例：页面在 https://example.com/todo/login → 返回 https://example.com/todo/
 */
export function getBaseRedirectUri(): string {
  const { origin, pathname } = window.location
  // 取到目录级别（去掉文件名部分），保留路径前缀
  const base = pathname.endsWith('/') ? pathname : pathname.substring(0, pathname.lastIndexOf('/') + 1)
  return origin + base
}
