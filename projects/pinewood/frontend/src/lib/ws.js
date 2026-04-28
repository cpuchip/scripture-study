// Reconnecting WebSocket with exponential backoff.
import { ref } from 'vue'

export function useSocket(onMessage) {
  const status = ref('connecting')
  let ws, retry = 0, closedByUs = false

  function connect() {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws = new WebSocket(`${proto}//${location.host}/ws`)
    ws.onopen = () => { status.value = 'open'; retry = 0 }
    ws.onmessage = (ev) => {
      try { onMessage(JSON.parse(ev.data)) } catch (e) { /* ignore */ }
    }
    ws.onclose = () => {
      status.value = 'closed'
      if (closedByUs) return
      const delay = Math.min(30000, 500 * 2 ** retry++)
      setTimeout(connect, delay)
    }
    ws.onerror = () => ws.close()
  }
  connect()

  return {
    status,
    close: () => { closedByUs = true; ws && ws.close() }
  }
}
