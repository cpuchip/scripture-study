// Service worker for Web Push notifications
// This file lives in /public/ and is served at the root of the domain.

self.addEventListener('push', function (event) {
  if (!event.data) return

  let payload
  try {
    payload = event.data.json()
  } catch {
    payload = {
      title: 'I Become',
      body: event.data.text(),
      url: '/today',
    }
  }

  const options = {
    body: payload.body || '',
    icon: '/ibecome-icon.png',
    badge: '/ibecome-icon.png',
    tag: payload.tag || 'default',
    renotify: true,
    data: {
      url: payload.url || '/today',
    },
  }

  event.waitUntil(self.registration.showNotification(payload.title || 'I Become', options))
})

self.addEventListener('notificationclick', function (event) {
  event.notification.close()

  const url = event.notification.data?.url || '/today'

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then(function (clientList) {
      // Focus existing tab if one is open
      for (const client of clientList) {
        if (client.url.includes(self.location.origin) && 'focus' in client) {
          client.navigate(url)
          return client.focus()
        }
      }
      // Otherwise open a new window
      return clients.openWindow(url)
    }),
  )
})
