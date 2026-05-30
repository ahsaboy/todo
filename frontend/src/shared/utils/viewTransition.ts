type ApplyFn = () => void | Promise<void>

interface RevealOrigin {
  clientX: number
  clientY: number
}

const TRANSITION_DURATION = 800

/**
 * 主题切换的「圆形揭示」过渡：基于 View Transitions API，从 origin（点击处）
 * 以 clip-path: circle 由 0 半径扩展到覆盖整个视口，呈现新主题从点击点泼开的效果。
 *
 * 浏览器不支持 startViewTransition 或用户开启「减少动态效果」时，直接应用变更、不做动画。
 * 注意：WAAPI 动画不受 CSS 的 prefers-reduced-motion 规则约束，必须在此主动判断。
 */
export function revealThemeTransition(origin: RevealOrigin | undefined, apply: ApplyFn): void {
  const prefersReducedMotion =
    typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches

  if (typeof document === 'undefined' || typeof document.startViewTransition !== 'function' || prefersReducedMotion) {
    void apply()
    return
  }

  const x = origin?.clientX ?? window.innerWidth / 2
  const y = origin?.clientY ?? window.innerHeight / 2
  const endRadius = Math.hypot(
    Math.max(x, window.innerWidth - x),
    Math.max(y, window.innerHeight - y),
  )

  const transition = document.startViewTransition(() => {
    void apply()
  })

  void transition.ready.then(() => {
    document.documentElement.animate(
      {
        clipPath: [
          `circle(0px at ${x}px ${y}px)`,
          `circle(${endRadius}px at ${x}px ${y}px)`,
        ],
      },
      {
        duration: TRANSITION_DURATION,
        easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
        pseudoElement: '::view-transition-new(root)',
      },
    )
  })
}
