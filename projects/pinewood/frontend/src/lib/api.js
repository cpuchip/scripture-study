async function req(method, url, body, opts = {}) {
  const init = { method, headers: {} }
  if (body !== undefined && !(body instanceof FormData)) {
    init.headers['Content-Type'] = 'application/json'
    init.body = JSON.stringify(body)
  } else if (body instanceof FormData) {
    init.body = body
  }
  const res = await fetch(url, init)
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  if (res.status === 204) return null
  if (opts.raw) return res
  const ct = res.headers.get('content-type') || ''
  return ct.includes('json') ? res.json() : res.text()
}

export const api = {
  listRaces: () => req('GET', '/api/races'),
  createRace: (name, laneCount = 3, parentID = null) =>
    req('POST', '/api/races', { name, lane_count: laneCount, parent_id: parentID }),
  getRace: (id) => req('GET', `/api/races/${id}`),
  deleteRace: (id) => req('DELETE', `/api/races/${id}`),
  listCars: (id) => req('GET', `/api/races/${id}/cars`),
  addCar: (id, number, name) => req('POST', `/api/races/${id}/cars`, { number, name }),
  updateCar: (id, carID, number, name) => req('PUT', `/api/races/${id}/cars/${carID}`, { number, name }),
  deleteCar: (id, carID) => req('DELETE', `/api/races/${id}/cars/${carID}`),
  finalize: (id) => req('POST', `/api/races/${id}/finalize`),
  heats: (id) => req('GET', `/api/races/${id}/heats`),
  heat: (id, n) => req('GET', `/api/races/${id}/heats/${n}`),
  score: (id, heat, lane, place) => req('POST', `/api/races/${id}/heats/${heat}/score`, { lane, place }),
  state: (id) => req('GET', `/api/races/${id}/state`),
  standings: (id) => req('GET', `/api/races/${id}/standings`),
  runoff: (id, cars, name) => req('POST', `/api/races/${id}/runoff`, { cars, name }),
  exportURL: (id) => `/api/races/${id}/export`,
  import: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return req('POST', '/api/import', fd)
  }
}
