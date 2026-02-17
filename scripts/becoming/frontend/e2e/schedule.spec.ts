import { test, expect, type APIRequestContext } from '@playwright/test'

// Helper: format a date as YYYY-MM-DD
function formatDate(d: Date): string {
  return d.toISOString().slice(0, 10)
}

// Helper: add days to a date
function addDays(d: Date, n: number): Date {
  const result = new Date(d)
  result.setDate(result.getDate() + n)
  return result
}

// Helper: create a practice via API
async function createPractice(
  request: APIRequestContext,
  practice: { name: string; type: string; category: string; config: string },
) {
  const res = await request.post('/api/practices', { data: practice })
  expect(res.ok(), `createPractice failed: ${res.status()}`).toBeTruthy()
  return res.json()
}

// Helper: create a log via API
async function createLog(
  request: APIRequestContext,
  log: { practice_id: number; date: string },
) {
  const res = await request.post('/api/logs', { data: log })
  expect(res.ok(), `createLog failed: ${res.status()}`).toBeTruthy()
  return res.json()
}

// Helper: get daily summary via API
async function getDailySummary(request: APIRequestContext, date: string) {
  const res = await request.get(`/api/daily/${date}`)
  expect(res.ok()).toBeTruthy()
  return res.json()
}

// Helper: delete a practice via API (cleanup)
async function deletePractice(request: APIRequestContext, id: number) {
  await request.delete(`/api/practices/${id}`)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// API-level tests: verify schedule math is correct
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

test.describe('Schedule shift behavior (API)', () => {
  let practiceId: number

  test.afterEach(async ({ request }) => {
    if (practiceId) {
      await deletePractice(request, practiceId)
      practiceId = 0
    }
  })

  test('shift_on_early: completing early shifts next_due forward from completion date', async ({
    request,
  }) => {
    // Anchor = today, log today, off-day = today+2 (before next due at today+5)
    const today = new Date()
    const todayStr = formatDate(today)
    const offDay = formatDate(addDays(today, 2))  // off-day: today+2
    const originalDue = formatDate(addDays(today, 5))  // next due: today+5

    const config = JSON.stringify({
      schedule: {
        type: 'interval',
        interval_days: 5,
        anchor_date: todayStr,
        shift_on_early: true,
      },
    })
    const practice = await createPractice(request, {
      name: 'PW-Shift-API',
      type: 'scheduled',
      category: 'test',
      config,
    })
    practiceId = practice.id

    // Log today (anchor day) to establish schedule
    await createLog(request, { practice_id: practiceId, date: todayStr })

    // Before off-day completion: should NOT be due, next_due = today+5
    const beforeSummary = await getDailySummary(request, offDay)
    const before = beforeSummary.find(
      (s: { practice_id: number }) => s.practice_id === practiceId,
    )
    expect(before).toBeTruthy()
    expect(before.is_due).toBe(false)
    expect(before.next_due).toBe(originalDue)

    // Complete on off-day
    await createLog(request, { practice_id: practiceId, date: offDay })

    // After: next_due should shift to offDay+5
    const afterSummary = await getDailySummary(request, offDay)
    const after = afterSummary.find(
      (s: { practice_id: number }) => s.practice_id === practiceId,
    )
    expect(after).toBeTruthy()
    expect(after.is_due).toBe(false) // just completed
    const shiftedDue = formatDate(addDays(new Date(offDay + 'T12:00:00'), 5))
    expect(after.next_due).toBe(shiftedDue)
  })

  test('fixed schedule: completing early does NOT shift next_due', async ({ request }) => {
    // Anchor = today, log today, off-day = today+2, next due stays at today+5
    const today = new Date()
    const todayStr = formatDate(today)
    const offDay = formatDate(addDays(today, 2))
    const originalDue = formatDate(addDays(today, 5))

    const config = JSON.stringify({
      schedule: {
        type: 'interval',
        interval_days: 5,
        anchor_date: todayStr,
        shift_on_early: false,
      },
    })
    const practice = await createPractice(request, {
      name: 'PW-Fixed-API',
      type: 'scheduled',
      category: 'test',
      config,
    })
    practiceId = practice.id

    // Log today
    await createLog(request, { practice_id: practiceId, date: todayStr })

    // Complete early on off-day (today+2)
    await createLog(request, { practice_id: practiceId, date: offDay })

    // next_due should STILL be today+5 (fixed schedule doesn't shift)
    const summary = await getDailySummary(request, offDay)
    const item = summary.find(
      (s: { practice_id: number }) => s.practice_id === practiceId,
    )
    expect(item).toBeTruthy()
    expect(item.is_due).toBe(false)
    expect(item.next_due).toBe(originalDue) // NOT shifted
  })

  test('off-day items still appear in daily summary', async ({ request }) => {
    // Anchor = today, log today, off-day = today+2
    const today = new Date()
    const todayStr = formatDate(today)
    const offDay = formatDate(addDays(today, 2))

    const config = JSON.stringify({
      schedule: {
        type: 'interval',
        interval_days: 5,
        anchor_date: todayStr,
        shift_on_early: true,
      },
    })
    const practice = await createPractice(request, {
      name: 'PW-Appear-API',
      type: 'scheduled',
      category: 'test',
      config,
    })
    practiceId = practice.id

    await createLog(request, { practice_id: practiceId, date: todayStr })

    // Off-day items should still appear in the daily summary (is_due=false)
    const summary = await getDailySummary(request, offDay)
    const item = summary.find(
      (s: { practice_id: number }) => s.practice_id === practiceId,
    )
    expect(item).toBeTruthy()
    expect(item.is_due).toBe(false)
    expect(item.next_due).toBeTruthy()
  })
})

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// UI tests: verify the off-day checkbox is clickable
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

test.describe('Off-day checkbox UI', () => {
  let practiceId: number

  test.afterEach(async ({ request }) => {
    if (practiceId) {
      await deletePractice(request, practiceId)
      practiceId = 0
    }
  })

  test('off-day scheduled practice shows clickable dashed checkbox', async ({
    page,
    request,
  }) => {
    // Set up: anchor = today - 2, interval = 5 → next due = today + 3
    // Today is an off-day (only 2 days since anchor, interval is 5)
    const today = new Date()
    const anchor = addDays(today, -2)
    const anchorStr = formatDate(anchor)

    const config = JSON.stringify({
      schedule: {
        type: 'interval',
        interval_days: 5,
        anchor_date: anchorStr,
        shift_on_early: true,
      },
    })
    const practice = await createPractice(request, {
      name: 'PW-OffDayUI',
      type: 'scheduled',
      category: 'test',
      config,
    })
    practiceId = practice.id

    // Log on anchor day so there's a last_log reference
    await createLog(request, { practice_id: practiceId, date: anchorStr })

    // Navigate to Today page (today IS the off-day — no arrow clicking needed)
    // Skip onboarding guard
    await page.goto('/today')
    await page.evaluate(() => localStorage.setItem('onboarding_complete', 'true'))
    await page.goto('/today')
    await page.waitForLoadState('networkidle')

    // Find the practice row — the item wrapper has class 'px-4 py-2'
    const row = page.locator('.px-4.py-2').filter({ hasText: /PW-OffDayUI/ })
    await expect(row).toBeVisible({ timeout: 10000 })

    // The checkbox should be a clickable BUTTON (not a static span)
    const checkbox = row.locator('button').first()
    await expect(checkbox).toBeVisible()

    // Should have dashed border style (off-day indicator)
    await expect(checkbox).toHaveCSS('border-style', 'dashed')
  })

  test('clicking off-day checkbox completes the practice and turns green', async ({
    page,
    request,
  }) => {
    const today = new Date()
    const todayStr = formatDate(today)
    const anchor = addDays(today, -2)
    const anchorStr = formatDate(anchor)

    const config = JSON.stringify({
      schedule: {
        type: 'interval',
        interval_days: 5,
        anchor_date: anchorStr,
        shift_on_early: true,
      },
    })
    const practice = await createPractice(request, {
      name: 'PW-ClickUI',
      type: 'scheduled',
      category: 'test',
      config,
    })
    practiceId = practice.id

    // Log on anchor day
    await createLog(request, { practice_id: practiceId, date: anchorStr })

    // Go to Today page
    // Skip onboarding guard
    await page.goto('/today')
    await page.evaluate(() => localStorage.setItem('onboarding_complete', 'true'))
    await page.goto('/today')
    await page.waitForLoadState('networkidle')

    // Find and click the off-day checkbox
    const row = page.locator('.px-4.py-2').filter({ hasText: /PW-ClickUI/ })
    await expect(row).toBeVisible({ timeout: 10000 })

    const checkbox = row.locator('button').first()
    await expect(checkbox).toBeVisible()
    await checkbox.click()

    // After clicking, should turn green (completed) — verify the class is applied
    await expect(checkbox).toHaveAttribute('class', /bg-green-500/, { timeout: 5000 })

    // Verify via API that a log was created
    const summary = await getDailySummary(request, todayStr)
    const item = summary.find(
      (s: { practice_id: number }) => s.practice_id === practiceId,
    )
    expect(item).toBeTruthy()
    expect(item.log_count).toBeGreaterThanOrEqual(1)
  })
})
