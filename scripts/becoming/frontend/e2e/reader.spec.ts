import { test, expect, type Page } from '@playwright/test'

// The public reader test uses the real repo as a test source
// This URL format matches the share page query params
const PUBLIC_READER_URL = '/share?p=gh&r=cpuchip/scripture-study&d=public/**/*.md'
const PUBLIC_FILE_URL = '/share?p=gh&r=cpuchip/scripture-study&d=public/**/*.md&f=public/study/word.md'

// Helper: wait for the file tree to finish loading by looking for a tree node entry
async function waitForFileTree(page: import('@playwright/test').Page) {
  // Tree nodes use span.font-medium inside cursor-pointer divs
  await page.locator('.reader-sidebar .overflow-y-auto span.font-medium').first().waitFor({ timeout: 15000 })
}

test.describe('Public Reader', () => {
  test('loads and displays file tree sidebar', async ({ page }) => {
    await page.goto(PUBLIC_READER_URL)

    // Wait for sidebar to load
    await expect(page.locator('.reader-sidebar')).toBeVisible()

    // Should show the repo name
    await expect(page.locator('.reader-sidebar h2')).toContainText('scripture-study')

    // Should have file tree entries (look for tree node spans, not heading text)
    await waitForFileTree(page)
    const treeNodes = page.locator('.reader-sidebar .overflow-y-auto span.font-medium')
    expect(await treeNodes.count()).toBeGreaterThan(0)
  })

  test('opens a document when clicking a file', async ({ page }) => {
    await page.goto(PUBLIC_READER_URL)

    // Wait for tree to load
    await waitForFileTree(page)

    // Expand the "study" folder to reveal files
    const studyFolder = page.locator('.reader-sidebar .overflow-y-auto [class*="cursor-pointer"]').filter({ hasText: /^study$/ }).first()
    if (await studyFolder.count() > 0) {
      await studyFolder.click()
    }

    // Click on a study document
    const fileEntry = page.locator('.reader-sidebar .overflow-y-auto [class*="cursor-pointer"]').filter({ hasText: 'word' }).first()
    await fileEntry.waitFor({ timeout: 5000 })
    await fileEntry.click()

    // Document should render
    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 10000 })
    await expect(page.locator('.reader-document article')).not.toBeEmpty()
  })

  test('opens directly to a specific file via URL param', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    // Should auto-load the document
    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })
    await expect(page.locator('.reader-document article')).toContainText('Word', { timeout: 10000 })
  })

  test('filter files via search box', async ({ page }) => {
    await page.goto(PUBLIC_READER_URL)

    await expect(page.locator('.reader-sidebar')).toBeVisible()
    await page.locator('.reader-sidebar input[placeholder="Filter files..."]').waitFor({ timeout: 10000 })

    // Wait for tree to actually load before filtering
    await waitForFileTree(page)

    await page.locator('.reader-sidebar input[placeholder="Filter files..."]').fill('word')

    // Should filter — file nodes use span.truncate for their name
    await expect(
      page.locator('.reader-sidebar .overflow-y-auto span.truncate').filter({ hasText: 'word' }).first()
    ).toBeVisible({ timeout: 5000 })
  })

  test('sidebar can be collapsed and expanded', async ({ page }) => {
    await page.goto(PUBLIC_READER_URL)

    await expect(page.locator('.reader-sidebar')).toBeVisible()

    // Click hide sidebar button
    await page.locator('button[title="Hide sidebar"]').click()
    await expect(page.locator('.reader-sidebar')).not.toBeVisible()

    // Click show sidebar button
    await page.locator('button[title="Show sidebar"]').click()
    await expect(page.locator('.reader-sidebar')).toBeVisible()
  })
})

test.describe('Dark Mode', () => {
  test('toggles dark mode on public reader', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })

    // Should start in light mode (no .dark class)
    await expect(page.locator('.public-reader')).not.toHaveClass(/dark/)

    // Click dark mode toggle
    await page.locator('button[title="Dark mode"]').click()

    // Should now have dark class
    await expect(page.locator('.public-reader')).toHaveClass(/dark/)

    // Persisted in localStorage
    const darkVal = await page.evaluate(() => localStorage.getItem('reader-dark-mode'))
    expect(darkVal).toBe('true')

    // Toggle back to light
    await page.locator('button[title="Light mode"]').click()
    await expect(page.locator('.public-reader')).not.toHaveClass(/dark/)
  })

  test('dark mode persists across page reload', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)
    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })

    // Enable dark mode
    await page.locator('button[title="Dark mode"]').click()
    await expect(page.locator('.public-reader')).toHaveClass(/dark/)

    // Reload
    await page.reload()
    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })

    // Should still be dark
    await expect(page.locator('.public-reader')).toHaveClass(/dark/)

    // Clean up
    await page.evaluate(() => localStorage.removeItem('reader-dark-mode'))
  })
})

test.describe('Reference Panel', () => {
  test('scripture links open in reference panel', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    // Wait for document to load
    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })
    await expect(page.locator('.reader-document article')).not.toBeEmpty({ timeout: 10000 })

    // Find a scripture reference link (data-ref-link or data-scripture-link)
    const refLink = page.locator('.reader-document a[data-ref-link], .reader-document a[data-scripture-link]').first()

    // If scripture links exist in the rendered content
    const linkCount = await refLink.count()
    if (linkCount > 0) {
      await refLink.click()

      // Reference panel should open
      await expect(page.locator('.reference-panel')).toBeVisible({ timeout: 10000 })

      // Panel should show content, loading, or error — any means it responded
      await expect(
        page.locator('.reference-panel .ref-content')
      ).toBeVisible({ timeout: 15000 })
    } else {
      // Skip if no scripture links in this document
      test.skip()
    }
  })

  test('reference panel tabs work', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })
    await expect(page.locator('.reader-document article')).not.toBeEmpty({ timeout: 10000 })

    const refLinks = page.locator('.reader-document a[data-ref-link], .reader-document a[data-scripture-link]')
    const count = await refLinks.count()

    if (count < 2) {
      test.skip()
      return
    }

    // Open first reference
    await refLinks.nth(0).click()
    await expect(page.locator('.reference-panel')).toBeVisible({ timeout: 10000 })

    // Wait for first tab content to load
    await page.waitForTimeout(2000)

    // Open second reference
    await refLinks.nth(1).click()

    // Should now have tab bar
    await expect(page.locator('.ref-tabs')).toBeVisible({ timeout: 5000 })

    // Both tabs should be visible
    const tabs = page.locator('.ref-tab')
    await expect(tabs).toHaveCount(2, { timeout: 5000 })
  })

  test('reference panel can be closed', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })
    await expect(page.locator('.reader-document article')).not.toBeEmpty({ timeout: 10000 })

    const refLink = page.locator('.reader-document a[data-ref-link], .reader-document a[data-scripture-link]').first()
    const linkCount = await refLink.count()

    if (linkCount === 0) {
      test.skip()
      return
    }

    await refLink.click()
    await expect(page.locator('.reference-panel')).toBeVisible({ timeout: 10000 })

    // Close the panel
    await page.locator('.reference-panel button[title="Close panel"]').click()
    await expect(page.locator('.reference-panel')).not.toBeVisible()
  })
})

test.describe('Mobile Responsive', () => {
  test.use({ viewport: { width: 375, height: 812 } }) // iPhone X

  test('sidebar auto-closes on mobile after file select', async ({ page }) => {
    await page.goto(PUBLIC_READER_URL)

    // On mobile, sidebar may start open or closed depending on content
    // Open it if needed
    const sidebarBtn = page.locator('button[title="Show sidebar"]')
    if (await sidebarBtn.isVisible()) {
      await sidebarBtn.click()
    }

    await expect(page.locator('.reader-sidebar')).toBeVisible({ timeout: 10000 })

    // Wait for tree to actually load
    await waitForFileTree(page)

    // Expand "study" folder to reveal files
    const studyFolder = page.locator('.reader-sidebar .overflow-y-auto [class*="cursor-pointer"]').filter({ hasText: /^study$/ }).first()
    if (await studyFolder.count() > 0) {
      await studyFolder.click()
      await page.waitForTimeout(300)
    }

    // Look for a file entry and click it
    const fileEntry = page.locator('.reader-sidebar .overflow-y-auto [class*="cursor-pointer"]').filter({ hasText: /word|creation|intelligence/i }).first()
    if (await fileEntry.count() === 0) {
      test.skip()
      return
    }

    await fileEntry.click()

    // Sidebar should auto-close on mobile
    await expect(page.locator('.reader-sidebar')).not.toBeVisible({ timeout: 5000 })

    // Document should be visible
    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 10000 })
  })

  test('mobile sidebar has overlay', async ({ page }) => {
    await page.goto(PUBLIC_READER_URL)

    // Open sidebar
    const sidebarBtn = page.locator('button[title="Show sidebar"]')
    if (await sidebarBtn.isVisible()) {
      await sidebarBtn.click()
    }

    await expect(page.locator('.reader-sidebar')).toBeVisible({ timeout: 10000 })

    // Should have overlay behind sidebar
    await expect(page.locator('.sidebar-overlay')).toBeVisible()

    // Click overlay to close — click to the right of the 280px sidebar
    // The overlay spans the full viewport, but the sidebar sits on top at left
    await page.locator('.sidebar-overlay').click({ position: { x: 350, y: 400 } })
    await expect(page.locator('.reader-sidebar')).not.toBeVisible()
  })

  test('reference panel goes full-width on mobile', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })
    await expect(page.locator('.reader-document article')).not.toBeEmpty({ timeout: 10000 })

    // On mobile the sidebar may be open and covering content — close it first
    const sidebar = page.locator('.reader-sidebar')
    if (await sidebar.isVisible()) {
      await page.locator('button[title="Hide sidebar"]').click()
      await expect(sidebar).not.toBeVisible()
    }

    const refLink = page.locator('.reader-document a[data-ref-link], .reader-document a[data-scripture-link]').first()
    if (await refLink.count() === 0) {
      test.skip()
      return
    }

    await refLink.click()
    await expect(page.locator('.reference-panel')).toBeVisible({ timeout: 10000 })

    // Panel should be full-width on mobile
    const panel = page.locator('.reference-panel')
    const box = await panel.boundingBox()
    expect(box).toBeTruthy()
    // Full width means it should span the viewport width
    expect(box!.width).toBeGreaterThanOrEqual(370)
  })
})

test.describe('Share Workflow (Public)', () => {
  test('share button copies link', async ({ page }) => {
    await page.goto(PUBLIC_FILE_URL)

    await expect(page.locator('.reader-document')).toBeVisible({ timeout: 15000 })

    // Grant clipboard permission
    await page.context().grantPermissions(['clipboard-read', 'clipboard-write'])

    // Click share button in header
    const shareBtn = page.locator('.public-header button').filter({ hasText: 'Share' })
    await shareBtn.click()

    // Clipboard should contain a share URL
    const clipText = await page.evaluate(() => navigator.clipboard.readText())
    expect(clipText).toContain('/share?')
  })
})

test.describe('Landing Page', () => {
  test('landing page loads', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/Become/i)
  })
})
