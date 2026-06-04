import { ref, onMounted, onBeforeUnmount, type Ref } from 'vue'

interface ParticleNetworkOptions {
  particleCount?: number
  connectionDistance?: number
  repulseRadius?: number
  repulseStrength?: number
  particleColor?: string
  lineColor?: string
  radiusRange?: [number, number]
  speedRange?: [number, number]
}

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  radius: number
  baseX: number
  baseY: number
}

export function useParticleNetwork(
  canvasRef: Ref<HTMLCanvasElement | null>,
  options: ParticleNetworkOptions = {}
) {
  const {
    particleCount = 70,
    connectionDistance = 120,
    repulseRadius = 150,
    repulseStrength = 0.08,
    particleColor = 'rgba(56, 217, 196, 0.6)',
    lineColor = 'rgba(56, 217, 196, 0.12)',
    radiusRange = [1, 3],
    speedRange = [0.2, 0.8],
  } = options

  const particleCountRef = ref(particleCount)
  const particleRepulseRadius = 50
  const particleRepulseStrength = 0.03
  const edgeRepulseRadius = 80
  const edgeRepulseStrength = 0.12
  let canvas: HTMLCanvasElement | null = null
  let ctx: CanvasRenderingContext2D | null = null
  let particles: Particle[] = []
  let mouse = { x: -9999, y: -9999 }
  let rafId: number | null = null
  let isIntersectionVisible = true
  let prefersReducedMotion = false

  const initCanvas = () => {
    if (!canvasRef.value) return

    canvas = canvasRef.value
    const parent = canvas.parentElement
    if (!parent) return

    ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = window.devicePixelRatio || 1
    const rect = parent.getBoundingClientRect()

    canvas.width = rect.width * dpr
    canvas.height = rect.height * dpr
    canvas.style.width = `${rect.width}px`
    canvas.style.height = `${rect.height}px`

    ctx.scale(dpr, dpr)

    initParticles(rect.width, rect.height)
  }

  const initParticles = (width: number, height: number) => {
    const count = prefersReducedMotion
      ? Math.ceil(particleCount / 2)
      : particleCount
    particleCountRef.value = count

    particles = Array.from({ length: count }, () => {
      const radius =
        radiusRange[0] +
        Math.random() * (radiusRange[1] - radiusRange[0])
      const speed =
        speedRange[0] +
        Math.random() * (speedRange[1] - speedRange[0])
      const angle = Math.random() * Math.PI * 2

      return {
        x: Math.random() * width,
        y: Math.random() * height,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed,
        radius,
        baseX: 0,
        baseY: 0,
      }
    })
  }

  const tick = () => {
    if (!canvas || !ctx || !isIntersectionVisible) {
      rafId = requestAnimationFrame(tick)
      return
    }

    const parent = canvas.parentElement
    if (!parent) {
      rafId = requestAnimationFrame(tick)
      return
    }

    const rect = parent.getBoundingClientRect()
    const width = rect.width
    const height = rect.height

    ctx.clearRect(0, 0, width, height)

    // Apply forces: mouse repulsion + particle repulsion + edge repulsion
    for (let i = 0; i < particles.length; i++) {
      const p = particles[i]

      // Mouse attraction
      if (mouse.x >= 0 && mouse.y >= 0) {
        const dx = mouse.x - p.x
        const dy = mouse.y - p.y
        const dist = Math.sqrt(dx * dx + dy * dy)

        if (dist < repulseRadius && dist > 0) {
          const force = (1 - dist / repulseRadius) * repulseStrength
          p.vx += (dx / dist) * force
          p.vy += (dy / dist) * force
        }
      }

      // Particle-to-particle repulsion
      for (let j = i + 1; j < particles.length; j++) {
        const p2 = particles[j]
        const dx = p.x - p2.x
        const dy = p.y - p2.y
        const dist = Math.sqrt(dx * dx + dy * dy)

        if (dist < particleRepulseRadius && dist > 0) {
          const force = (1 - dist / particleRepulseRadius) * particleRepulseStrength
          const fx = (dx / dist) * force
          const fy = (dy / dist) * force

          p.vx += fx
          p.vy += fy
          p2.vx -= fx
          p2.vy -= fy
        }
      }

      // Edge repulsion (push away from edges)
      const edgeDistances = [
        { dist: p.x, nx: 1, ny: 0 },              // left edge → push right
        { dist: width - p.x, nx: -1, ny: 0 },     // right edge → push left
        { dist: p.y, nx: 0, ny: 1 },              // top edge → push down
        { dist: height - p.y, nx: 0, ny: -1 },    // bottom edge → push up
      ]

      for (const edge of edgeDistances) {
        if (edge.dist < edgeRepulseRadius) {
          const force = (1 - edge.dist / edgeRepulseRadius) * edgeRepulseStrength
          p.vx += edge.nx * force
          p.vy += edge.ny * force
        }
      }
    }

    // Update positions
    for (const p of particles) {
      p.x += p.vx
      p.y += p.vy

      // Boundary bounce with damping
      if (p.x < 0 || p.x > width) {
        p.vx *= -0.8
        p.x = Math.max(0, Math.min(width, p.x))
      }
      if (p.y < 0 || p.y > height) {
        p.vy *= -0.8
        p.y = Math.max(0, Math.min(height, p.y))
      }

      // Damping to return to normal speed
      p.vx *= 0.99
      p.vy *= 0.99
    }

    // Draw connections
    if (!prefersReducedMotion) {
      ctx.strokeStyle = lineColor
      ctx.lineWidth = 0.5

      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const p1 = particles[i]
          const p2 = particles[j]
          const dx = p1.x - p2.x
          const dy = p1.y - p2.y
          const dist = Math.sqrt(dx * dx + dy * dy)

          if (dist < connectionDistance) {
            const opacity = 1 - dist / connectionDistance
            ctx.globalAlpha = opacity
            ctx.beginPath()
            ctx.moveTo(p1.x, p1.y)
            ctx.lineTo(p2.x, p2.y)
            ctx.stroke()
          }
        }
      }
      ctx.globalAlpha = 1
    }

    // Draw particles
    for (const p of particles) {
      ctx.beginPath()
      ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2)
      ctx.fillStyle = particleColor
      ctx.fill()
    }

    rafId = requestAnimationFrame(tick)
  }

  const handleMouseMove = (e: MouseEvent) => {
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    mouse.x = e.clientX - rect.left
    mouse.y = e.clientY - rect.top
  }

  const handleMouseLeave = () => {
    mouse.x = -9999
    mouse.y = -9999
  }

  const restart = () => {
    if (rafId) {
      cancelAnimationFrame(rafId)
      rafId = null
    }
    initCanvas()
    rafId = requestAnimationFrame(tick)
  }

  onMounted(() => {
    // Check prefers-reduced-motion
    const motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    prefersReducedMotion = motionQuery.matches

    // Initialize canvas
    initCanvas()
    rafId = requestAnimationFrame(tick)

    // Mouse events on document (page-wide attraction)
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseleave', handleMouseLeave)

    // ResizeObserver
    const parent = canvasRef.value?.parentElement
    if (parent) {
      const resizeObserver = new ResizeObserver(() => {
        initCanvas()
      })
      resizeObserver.observe(parent)
    }

    // IntersectionObserver
    const intersectionObserver = new IntersectionObserver(
      (entries) => {
        isIntersectionVisible = entries[0]?.isIntersecting ?? true
      },
      { threshold: 0 }
    )

    if (canvas) {
      intersectionObserver.observe(canvas)
    }
  })

  onBeforeUnmount(() => {
    if (rafId) {
      cancelAnimationFrame(rafId)
      rafId = null
    }

    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseleave', handleMouseLeave)

    canvas = null
    ctx = null
    particles = []
  })

  return {
    restart,
    particleCount: particleCountRef,
  }
}
