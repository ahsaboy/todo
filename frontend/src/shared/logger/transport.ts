import { API_BASE_URL } from '@/shared/config/api'
import type { LogEntry } from './types'

const MAX_QUEUE_SIZE = 20

let queue: LogEntry[] = []
let flushScheduled = false
let unloadListenerRegistered = false
let isFlushing = false
const FLUSH_DELAY_MS = 50

function createPayload(entries: LogEntry[]): string {
  return JSON.stringify(entries)
}

function flushWithBeacon(entries: LogEntry[]): boolean {
  if (!navigator.sendBeacon) {
    return false
  }

  const blob = new Blob([createPayload(entries)], { type: 'application/json' })
  return navigator.sendBeacon(`${API_BASE_URL}/logs/frontend`, blob)
}

async function flushWithFetch(entries: LogEntry[]): Promise<void> {
  await fetch(`${API_BASE_URL}/logs/frontend`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: createPayload(entries),
    keepalive: true,
  })
}

async function flushQueue(): Promise<void> {
  if (isFlushing || queue.length === 0) {
    return
  }

  isFlushing = true
  const entries = queue
  queue = []

  if (!flushWithBeacon(entries)) {
    try {
      await flushWithFetch(entries)
    } catch {
      // Intentionally ignore transport errors.
    }
  }
  isFlushing = false

  if (queue.length > 0) {
    scheduleFlush()
  }
}

function scheduleFlush(): void {
  if (flushScheduled) {
    return
  }
  flushScheduled = true
  setTimeout(() => {
    flushScheduled = false
    void flushQueue()
  }, FLUSH_DELAY_MS)
}

function ensureUnloadListener(): void {
  if (unloadListenerRegistered) {
    return
  }
  unloadListenerRegistered = true
  const handleFlush = () => {
    void flushQueue()
  }
  window.addEventListener('pagehide', handleFlush)
  window.addEventListener('unload', handleFlush)
}

export function enqueueLog(entry: LogEntry): void {
  queue.push(entry)
  if (queue.length > MAX_QUEUE_SIZE) {
    queue = queue.slice(queue.length - MAX_QUEUE_SIZE)
  }
  ensureUnloadListener()
  scheduleFlush()
}

export async function flushLogs(): Promise<void> {
  await flushQueue()
}
