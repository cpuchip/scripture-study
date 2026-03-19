import { ref } from 'vue'
import { request } from '../api'

export type NotificationPermission = 'default' | 'granted' | 'denied'

const permission = ref<NotificationPermission>(
  typeof Notification !== 'undefined' ? Notification.permission : 'default',
)
const subscribed = ref(false)
const loading = ref(false)
const supported = ref('serviceWorker' in navigator && 'PushManager' in window)

export function useNotifications() {
  async function getVAPIDKey(): Promise<string | null> {
    try {
      const data = await request<{ public_key: string }>('/push/vapid-key')
      return data.public_key
    } catch {
      return null
    }
  }

  async function subscribe(): Promise<boolean> {
    if (!supported.value) return false
    loading.value = true

    try {
      // Request notification permission
      const perm = await Notification.requestPermission()
      permission.value = perm
      if (perm !== 'granted') {
        return false
      }

      // Get VAPID key from server
      const vapidKey = await getVAPIDKey()
      if (!vapidKey) {
        console.error('Failed to get VAPID key')
        return false
      }

      // Get service worker registration
      const registration = await navigator.serviceWorker.ready

      // Subscribe to push
      const pushSubscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(vapidKey) as BufferSource,
      })

      // Send subscription to server
      const sub = pushSubscription.toJSON()
      await request('/push/subscribe', {
        method: 'POST',
        body: JSON.stringify({
          endpoint: sub.endpoint,
          keys: {
            p256dh: sub.keys?.p256dh,
            auth: sub.keys?.auth,
          },
        }),
      })

      subscribed.value = true
      return true
    } catch (err) {
      console.error('Failed to subscribe to push notifications:', err)
      return false
    } finally {
      loading.value = false
    }
  }

  async function unsubscribe(): Promise<boolean> {
    if (!supported.value) return false
    loading.value = true

    try {
      const registration = await navigator.serviceWorker.ready
      const pushSubscription = await registration.pushManager.getSubscription()

      if (pushSubscription) {
        // Tell server to remove this subscription
        await request('/push/unsubscribe', {
          method: 'DELETE',
          body: JSON.stringify({
            endpoint: pushSubscription.endpoint,
          }),
        })

        // Unsubscribe locally
        await pushSubscription.unsubscribe()
      }

      subscribed.value = false
      return true
    } catch (err) {
      console.error('Failed to unsubscribe from push notifications:', err)
      return false
    } finally {
      loading.value = false
    }
  }

  async function checkSubscription(): Promise<void> {
    if (!supported.value) return
    try {
      const registration = await navigator.serviceWorker.ready
      const pushSubscription = await registration.pushManager.getSubscription()
      subscribed.value = pushSubscription !== null
      permission.value = Notification.permission
    } catch {
      subscribed.value = false
    }
  }

  async function sendTest(): Promise<boolean> {
    try {
      await request('/push/test', { method: 'POST' })
      return true
    } catch {
      return false
    }
  }

  return {
    permission,
    subscribed,
    loading,
    supported,
    subscribe,
    unsubscribe,
    checkSubscription,
    sendTest,
  }
}

// Convert a URL-safe base64 string to a Uint8Array (for applicationServerKey)
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)
  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i)
  }
  return outputArray
}
