// Package tui provides the terminal user interface for Gospel Library Downloader.
package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/api"
	"github.com/cpuchip/scripture-study/scripts/gospel-library/internal/cache"
)

// State represents the current view state of the TUI.
type State int

const (
	StateLoading State = iota
	StateBrowsing
	StateError
	StateDownloading
	StateConverting
)

// NavItem represents a navigable item in the list.
type NavItem struct {
	title    string
	desc     string
	uri      string
	itemType string // "collection", "toc", "content", "section"
	cached   bool
}

func (i NavItem) Title() string       { return i.title }
func (i NavItem) Description() string { return i.desc }
func (i NavItem) FilterValue() string { return i.title }

// Model is the main TUI model.
type Model struct {
	// Core components
	client    *cache.CachedClient
	rawClient *api.Client
	fileCache *cache.Cache
	lang      string
	outputDir string

	// UI state
	state     State
	list      list.Model
	spinner   spinner.Model
	width     int
	height    int
	err       error
	statusMsg string // Temporary status message

	// Navigation
	breadcrumbs  []NavItem
	currentURI   string
	currentTitle string

	// Selection for batch operations
	selected map[string]bool

	// Download
	downloader          *Downloader
	downloading         int // Total number of items to download
	downloadingDone     int // Number of items completed
	downloadQueue       []string
	downloadResults     []DownloadResult
	currentDownloadURI  string
	currentDownloadPath string
	lastDownloadStatus  string

	// Crawl progress
	crawling     bool
	crawlVisited int
	crawlFound   int
	crawlCurrent string
	crawlMsgCh   chan tea.Msg
}

// New creates a new TUI model.
func New(client *cache.CachedClient, rawClient *api.Client, fileCache *cache.Cache, lang, outputDir string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create list with default delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderLeftForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("241")).
		BorderLeftForeground(lipgloss.Color("170"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Gospel Library"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	return Model{
		client:      client,
		rawClient:   rawClient,
		fileCache:   fileCache,
		lang:        lang,
		outputDir:   outputDir,
		state:       StateLoading,
		list:        l,
		spinner:     s,
		breadcrumbs: []NavItem{},
		selected:    make(map[string]bool),
		downloader:  NewDownloader(client, rawClient, lang, outputDir),
	}
}

// Keybindings
type keyMap struct {
	Enter       key.Binding
	Back        key.Binding
	Select      key.Binding
	SelectAll   key.Binding
	Download    key.Binding
	DownloadAll key.Binding
	ClearSelect key.Binding
	Quit        key.Binding
	Help        key.Binding
}

var keys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "browse into"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace", "esc"),
		key.WithHelp("backspace", "go back"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle select"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle all"),
	),
	Download: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "download selected"),
	),
	DownloadAll: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "download view"),
	),
	ClearSelect: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear selection"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// Messages
type errMsg struct{ err error }
type loadedMsg struct {
	items []NavItem
	title string
}
type crawlProgressMsg struct {
	current    string
	visited    int
	discovered int
}
type crawlCompleteMsg struct {
	uris  []string
	title string
}

func (e errMsg) Error() string { return e.err.Error() }

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadRoot(),
	)
}

// loadRoot loads the root navigation (main categories).
func (m Model) loadRoot() tea.Cmd {
	return func() tea.Msg {
		items := []NavItem{
			{title: "General Conference", desc: "Conference talks from prophets and apostles", uri: "/general-conference", itemType: "collection"},
			{title: "Scriptures", desc: "Standard works of the Church", uri: "/scriptures", itemType: "collection"},
			{title: "Come, Follow Me", desc: "Home-centered study curriculum", uri: "/come-follow-me", itemType: "collection"},
			{title: "Manuals", desc: "Handbooks and callings resources", uri: "/handbooks-and-callings", itemType: "collection"},
			{title: "Magazines", desc: "Ensign, Liahona, and more", uri: "/magazines", itemType: "collection"},
		}
		return loadedMsg{items: items, title: "Gospel Library"}
	}
}

// loadCollection loads a collection from the API.
func (m Model) loadCollection(uri string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Try to get as a collection first
		collection, _, err := m.client.GetCollection(ctx, uri)
		if err == nil && collection != nil {
			var items []NavItem
			for _, section := range collection.Sections {
				for _, entry := range section.Entries {
					itemType := "collection"
					if entry.Type == "item" {
						if isContentURI(entry.URI) {
							itemType = "content"
						}
					}
					desc := entry.Category
					if section.Title != "" && section.Title != collection.Title {
						desc = section.Title
					}
					items = append(items, NavItem{
						title:    entry.Title,
						desc:     desc,
						uri:      entry.URI,
						itemType: itemType,
						cached:   m.fileCache.Has(entry.URI, "content"),
					})
				}
			}
			return loadedMsg{items: items, title: collection.Title}
		}

		// Try dynamic endpoint (for conference sessions, etc.)
		dynamic, _, err := m.client.GetDynamic(ctx, uri)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to load %s: %w", uri, err)}
		}

		var items []NavItem
		var title string

		if dynamic.TOC != nil {
			title = dynamic.TOC.Title
			for _, entry := range dynamic.TOC.Entries {
				if entry.Content != nil {
					itemType := "content"
					if !isContentURI(entry.Content.URI) {
						itemType = "collection"
					}
					items = append(items, NavItem{
						title:    entry.Content.Title,
						desc:     "Talk",
						uri:      entry.Content.URI,
						itemType: itemType,
						cached:   m.fileCache.Has(entry.Content.URI, "content"),
					})
				}
				if entry.Section != nil {
					// Add section as navigable item
					items = append(items, NavItem{
						title:    entry.Section.Title,
						desc:     fmt.Sprintf("%d items", len(entry.Section.Entries)),
						uri:      entry.Section.URI,
						itemType: "section",
					})
				}
			}
		} else if dynamic.Collection != nil {
			title = dynamic.Collection.Title
			for _, section := range dynamic.Collection.Sections {
				for _, entry := range section.Entries {
					itemType := "collection"
					if entry.Type == "item" {
						if isContentURI(entry.URI) {
							itemType = "content"
						}
					}
					items = append(items, NavItem{
						title:    entry.Title,
						desc:     entry.Category,
						uri:      entry.URI,
						itemType: itemType,
						cached:   m.fileCache.Has(entry.URI, "content"),
					})
				}
			}
		}

		if len(items) == 0 {
			return errMsg{err: fmt.Errorf("no items found at %s", uri)}
		}

		return loadedMsg{items: items, title: title}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Back):
			if len(m.breadcrumbs) > 0 {
				// Go back
				prev := m.breadcrumbs[len(m.breadcrumbs)-1]
				m.breadcrumbs = m.breadcrumbs[:len(m.breadcrumbs)-1]
				m.state = StateLoading
				m.currentURI = prev.uri
				if len(m.breadcrumbs) == 0 {
					return m, tea.Batch(m.spinner.Tick, m.loadRoot())
				}
				return m, tea.Batch(m.spinner.Tick, m.loadCollection(prev.uri))
			}

		case key.Matches(msg, keys.Enter):
			if m.state == StateBrowsing {
				if item, ok := m.list.SelectedItem().(NavItem); ok {
					uri := m.getOriginalURI(item.uri)
					if item.itemType == "content" {
						// For content items, toggle selection (can't browse deeper)
						if m.selected[uri] {
							delete(m.selected, uri)
						} else {
							m.selected[uri] = true
						}
						return m, m.refreshCurrentView()
					}
					// Navigate into collection/section
					m.breadcrumbs = append(m.breadcrumbs, NavItem{
						title: m.currentTitle,
						uri:   m.currentURI,
					})
					m.currentURI = uri
					m.state = StateLoading
					return m, tea.Batch(m.spinner.Tick, m.loadCollection(uri))
				}
			}

		case key.Matches(msg, keys.Select):
			if m.state == StateBrowsing {
				if item, ok := m.list.SelectedItem().(NavItem); ok {
					// Get the original URI without prefixes
					uri := m.getOriginalURI(item.uri)
					// Toggle selection for ANY item type (content, collection, or section)
					if m.selected[uri] {
						delete(m.selected, uri)
					} else {
						m.selected[uri] = true
					}
					// Refresh list to show updated selection
					return m, m.refreshCurrentView()
				}
			}

		case key.Matches(msg, keys.Download):
			if m.state == StateBrowsing && len(m.selected) > 0 {
				// Separate content items from collections/sections
				var contentURIs []string
				var collectionItems []struct {
					uri   string
					title string
				}

				// Get the item types from the current list
				items := m.list.Items()
				itemTypes := make(map[string]string)
				itemTitles := make(map[string]string)
				for _, item := range items {
					if navItem, ok := item.(NavItem); ok {
						itemTypes[navItem.uri] = navItem.itemType
						itemTitles[navItem.uri] = navItem.title
					}
				}

				for uri := range m.selected {
					itemType := itemTypes[uri]
					if itemType == "content" {
						contentURIs = append(contentURIs, uri)
					} else if itemType == "collection" || itemType == "section" {
						collectionItems = append(collectionItems, struct {
							uri   string
							title string
						}{uri, itemTitles[uri]})
					}
				}

				// If we have collections, crawl them first
				if len(collectionItems) > 0 {
					m.state = StateDownloading
					m.statusMsg = fmt.Sprintf("Crawling %d collections...", len(collectionItems))
					m.crawling = true
					m.crawlVisited = 0
					m.crawlFound = len(contentURIs)
					m.crawlCurrent = ""
					m.crawlMsgCh = make(chan tea.Msg)
					return m, tea.Batch(
						m.spinner.Tick,
						m.crawlAndDownloadMultiple(collectionItems, contentURIs),
					)
				}

				// Otherwise just download content items
				if len(contentURIs) > 0 {
					m.state = StateDownloading
					m.statusMsg = fmt.Sprintf("Downloading %d items...", len(contentURIs))
					m, cmd := m.beginDownloadQueue(contentURIs)
					return m, cmd
				}
			} else if m.state == StateBrowsing && len(m.selected) == 0 {
				m.statusMsg = "No items selected. Use space to select items."
			}

		case key.Matches(msg, keys.SelectAll):
			if m.state == StateBrowsing {
				// Toggle all content items in current view
				items := m.list.Items()
				// First check if all are already selected
				allSelected := true
				contentCount := 0
				for _, item := range items {
					if navItem, ok := item.(NavItem); ok && navItem.itemType == "content" {
						uri := m.getOriginalURI(navItem.uri)
						contentCount++
						if !m.selected[uri] {
							allSelected = false
						}
					}
				}
				// If all selected, deselect all; otherwise select all
				for _, item := range items {
					if navItem, ok := item.(NavItem); ok && navItem.itemType == "content" {
						uri := m.getOriginalURI(navItem.uri)
						if allSelected {
							delete(m.selected, uri)
						} else {
							m.selected[uri] = true
						}
					}
				}
				if allSelected {
					m.statusMsg = fmt.Sprintf("Deselected %d content items", contentCount)
				} else {
					m.statusMsg = fmt.Sprintf("Selected %d content items", contentCount)
				}
				return m, m.refreshCurrentView()
			}

		case key.Matches(msg, keys.DownloadAll):
			if m.state == StateBrowsing {
				// Download ALL content items in current view (ignores selection)
				items := m.list.Items()
				var uris []string
				for _, item := range items {
					if navItem, ok := item.(NavItem); ok && navItem.itemType == "content" {
						uri := m.getOriginalURI(navItem.uri)
						uris = append(uris, uri)
					}
				}
				if len(uris) > 0 {
					m.state = StateDownloading
					m.statusMsg = fmt.Sprintf("Downloading all %d items...", len(uris))
					m, cmd := m.beginDownloadQueue(uris)
					return m, cmd
				} else {
					m.statusMsg = "No content items to download. Navigate to a page with chapters/talks."
				}
			}

		case key.Matches(msg, keys.ClearSelect):
			if m.state == StateBrowsing && len(m.selected) > 0 {
				count := len(m.selected)
				m.selected = make(map[string]bool)
				m.statusMsg = fmt.Sprintf("Cleared %d selected items", count)
				return m, m.refreshCurrentView()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-3)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case loadedMsg:
		m.state = StateBrowsing
		m.currentTitle = msg.title
		m.list.Title = msg.title

		// Store current cursor position
		cursorIdx := m.list.Index()

		items := make([]list.Item, len(msg.items))
		for i, item := range msg.items {
			items[i] = m.formatNavItem(item)
		}
		m.list.SetItems(items)

		// Restore cursor position
		if cursorIdx < len(items) {
			m.list.Select(cursorIdx)
		}
		return m, nil

	case errMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case downloadCompleteMsg:
		// Downloads finished
		m.state = StateBrowsing
		successCount := 0
		for _, r := range msg.results {
			if r.Success {
				successCount++
			}
		}
		m.statusMsg = fmt.Sprintf("✓ Downloaded %d/%d items to %s/", successCount, len(msg.results), m.outputDir)
		m.downloadQueue = nil
		m.downloadResults = nil
		m.downloading = 0
		m.downloadingDone = 0
		m.currentDownloadURI = ""
		m.currentDownloadPath = ""
		m.lastDownloadStatus = ""
		// Clear selection
		m.selected = make(map[string]bool)
		return m, nil

	case downloadResultMsg:
		// Single download finished
		m.downloadResults = append(m.downloadResults, msg.result)
		m.downloadingDone++
		if msg.result.Success {
			m.lastDownloadStatus = fmt.Sprintf("✓ %s", msg.result.Title)
		} else {
			m.lastDownloadStatus = fmt.Sprintf("✗ %v", msg.result.Error)
		}
		if len(m.downloadQueue) > 0 {
			m.state = StateDownloading
			m, cmd := m.startNextDownload()
			return m, cmd
		}
		// All downloads complete
		return m, func() tea.Msg { return downloadCompleteMsg{results: m.downloadResults} }

	case crawlCompleteMsg:
		// Crawl finished, start downloads
		m.crawling = false
		m.crawlMsgCh = nil
		m.crawlCurrent = ""
		if len(msg.uris) == 0 {
			m.state = StateBrowsing
			m.statusMsg = fmt.Sprintf("No content found under %s", msg.title)
			return m, nil
		}
		m.state = StateDownloading
		m.statusMsg = fmt.Sprintf("Downloading %d items from %s...", len(msg.uris), msg.title)
		m, cmd := m.beginDownloadQueue(msg.uris)
		return m, cmd

	case crawlProgressMsg:
		m.crawlCurrent = msg.current
		m.crawlVisited = msg.visited
		m.crawlFound = msg.discovered
		m.statusMsg = fmt.Sprintf("Crawling... %d discovered", msg.discovered)
		return m, m.listenCrawl()
	}

	// Update list
	if m.state == StateBrowsing {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

// getOriginalURI strips visual prefixes from a URI that may have been added for display.
func (m Model) getOriginalURI(displayedURI string) string {
	// The NavItem stores the original URI, but titles may have prefixes
	// This function extracts the original URI from the item
	return displayedURI
}

// listenCrawl listens for the next crawl progress or completion message.
func (m Model) listenCrawl() tea.Cmd {
	ch := m.crawlMsgCh
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// buildOutputPath returns the target markdown path for a given URI.
func (m Model) buildOutputPath(uri string) string {
	cleanURI := strings.TrimPrefix(uri, "/")
	filename := filepath.Base(cleanURI) + ".md"
	dir := filepath.Dir(cleanURI)
	return filepath.Join(m.outputDir, m.lang, dir, filename)
}

// beginDownloadQueue initializes the download queue and starts the first download.
func (m Model) beginDownloadQueue(uris []string) (Model, tea.Cmd) {
	m.downloadQueue = append([]string{}, uris...)
	m.downloadResults = nil
	m.downloading = len(uris)
	m.downloadingDone = 0
	m.currentDownloadURI = ""
	m.currentDownloadPath = ""
	m.lastDownloadStatus = ""
	m, cmd := m.startNextDownload()
	return m, cmd
}

// startNextDownload kicks off the next download in the queue.
func (m Model) startNextDownload() (Model, tea.Cmd) {
	if len(m.downloadQueue) == 0 {
		return m, func() tea.Msg { return downloadCompleteMsg{results: m.downloadResults} }
	}

	uri := m.downloadQueue[0]
	m.downloadQueue = m.downloadQueue[1:]
	m.currentDownloadURI = uri
	m.currentDownloadPath = m.buildOutputPath(uri)
	m.statusMsg = fmt.Sprintf("Downloading %d/%d items...", m.downloadingDone+1, m.downloading)

	return m, tea.Batch(
		m.spinner.Tick,
		m.downloader.DownloadSingle(context.Background(), uri),
	)
}

// formatNavItem adds visual indicators to a NavItem for display.
// Status indicators (single column, left of title):
//   - "  " (blank) = not selected, not downloaded
//   - "● " = selected (pending download)
//   - "✓ " = downloaded/cached (not selected)
//   - "◉ " = selected AND cached (will re-download)
func (m Model) formatNavItem(item NavItem) NavItem {
	var prefix string

	isSelected := m.selected[item.uri]
	isCached := item.cached

	switch {
	case isSelected && isCached:
		prefix = "◉ " // Selected + cached (will re-download)
	case isSelected:
		prefix = "● " // Selected (pending download)
	case isCached:
		prefix = "✓ " // Cached (already downloaded)
	default:
		prefix = "  " // Neither
	}

	return NavItem{
		title:    prefix + item.title,
		desc:     item.desc,
		uri:      item.uri,
		itemType: item.itemType,
		cached:   item.cached,
	}
}

// refreshCurrentView reloads the current view to update visual state.
func (m Model) refreshCurrentView() tea.Cmd {
	return func() tea.Msg {
		// Re-fetch current items to refresh the display
		if len(m.breadcrumbs) == 0 && m.currentURI == "" {
			// At root
			items := []NavItem{
				{title: "General Conference", desc: "Conference talks from prophets and apostles", uri: "/general-conference", itemType: "collection"},
				{title: "Scriptures", desc: "Standard works of the Church", uri: "/scriptures", itemType: "collection"},
				{title: "Come, Follow Me", desc: "Home-centered study curriculum", uri: "/come-follow-me", itemType: "collection"},
				{title: "Manuals", desc: "Handbooks and callings resources", uri: "/handbooks-and-callings", itemType: "collection"},
				{title: "Magazines", desc: "Ensign, Liahona, and more", uri: "/magazines", itemType: "collection"},
			}
			return loadedMsg{items: items, title: "Gospel Library"}
		}

		// Otherwise reload current location
		ctx := context.Background()
		uri := m.currentURI

		// Try collection first
		collection, _, err := m.client.GetCollection(ctx, uri)
		if err == nil && collection != nil {
			var items []NavItem
			for _, section := range collection.Sections {
				for _, entry := range section.Entries {
					itemType := "collection"
					if entry.Type == "item" {
						if isContentURI(entry.URI) {
							itemType = "content"
						}
					}
					desc := entry.Category
					if section.Title != "" && section.Title != collection.Title {
						desc = section.Title
					}
					items = append(items, NavItem{
						title:    entry.Title,
						desc:     desc,
						uri:      entry.URI,
						itemType: itemType,
						cached:   m.fileCache.Has(entry.URI, "content"),
					})
				}
			}
			return loadedMsg{items: items, title: collection.Title}
		}

		// Try dynamic endpoint
		dynamic, _, err := m.client.GetDynamic(ctx, uri)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to refresh %s: %w", uri, err)}
		}

		var items []NavItem
		var title string

		if dynamic.TOC != nil {
			title = dynamic.TOC.Title
			for _, entry := range dynamic.TOC.Entries {
				if entry.Content != nil {
					itemType := "content"
					if !isContentURI(entry.Content.URI) {
						itemType = "collection"
					}
					items = append(items, NavItem{
						title:    entry.Content.Title,
						desc:     "Talk",
						uri:      entry.Content.URI,
						itemType: itemType,
						cached:   m.fileCache.Has(entry.Content.URI, "content"),
					})
				}
				if entry.Section != nil {
					items = append(items, NavItem{
						title:    entry.Section.Title,
						desc:     fmt.Sprintf("%d items", len(entry.Section.Entries)),
						uri:      entry.Section.URI,
						itemType: "section",
					})
				}
			}
		} else if dynamic.Collection != nil {
			title = dynamic.Collection.Title
			for _, section := range dynamic.Collection.Sections {
				for _, entry := range section.Entries {
					itemType := "collection"
					if entry.Type == "item" {
						if isContentURI(entry.URI) {
							itemType = "content"
						}
					}
					items = append(items, NavItem{
						title:    entry.Title,
						desc:     entry.Category,
						uri:      entry.URI,
						itemType: itemType,
						cached:   m.fileCache.Has(entry.URI, "content"),
					})
				}
			}
		}

		return loadedMsg{items: items, title: title}
	}
}

// crawlAndDownload recursively discovers all content under a URI and downloads them.
func (m Model) crawlAndDownload(uri, title string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Use the downloader's CrawlForContent which properly filters for actual content
		uris, err := m.downloader.CrawlForContent(ctx, uri)
		if err != nil {
			return errMsg{err: fmt.Errorf("crawl failed: %w", err)}
		}

		return crawlCompleteMsg{uris: uris, title: title}
	}
}

// crawlAndDownloadMultiple crawls multiple collections and merges with direct content URIs.
// collectionItems is a slice of {uri, title} for collections/sections to crawl recursively.
// contentURIs is a slice of direct content URIs to download.
func (m Model) crawlAndDownloadMultiple(collectionItems []struct{ uri, title string }, contentURIs []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		ch := m.crawlMsgCh
		if ch == nil {
			return errMsg{err: fmt.Errorf("crawl channel not initialized")}
		}

		// Start with direct content URIs
		allURIs := make([]string, len(contentURIs))
		copy(allURIs, contentURIs)

		go func() {
			defer close(ch)
			totalFound := len(contentURIs)
			totalVisited := 0

			for _, item := range collectionItems {
				// Send a progress update for the new root
				ch <- crawlProgressMsg{current: item.uri, visited: totalVisited, discovered: totalFound}

				uris, visited, err := m.downloader.CrawlForContentWithProgress(ctx, item.uri, func(current string, visitedCount int, discoveredCount int) {
					ch <- crawlProgressMsg{
						current:    current,
						visited:    totalVisited + visitedCount,
						discovered: totalFound + discoveredCount,
					}
				})
				if err != nil {
					continue
				}
				allURIs = append(allURIs, uris...)
				totalFound += len(uris)
				totalVisited += visited
			}

			// Remove duplicates
			seen := make(map[string]bool)
			uniqueURIs := make([]string, 0, len(allURIs))
			for _, uri := range allURIs {
				if !seen[uri] {
					seen[uri] = true
					uniqueURIs = append(uniqueURIs, uri)
				}
			}

			title := fmt.Sprintf("%d selected items", len(collectionItems)+len(contentURIs))
			ch <- crawlCompleteMsg{uris: uniqueURIs, title: title}
		}()

		return <-ch
	}
}

// View renders the UI.
func (m Model) View() string {
	switch m.state {
	case StateLoading:
		return loadingStyle.Render(fmt.Sprintf("%s Loading...", m.spinner.View()))

	case StateDownloading:
		if m.crawling {
			lines := []string{
				fmt.Sprintf("%s Crawling", m.spinner.View()),
				fmt.Sprintf("Discovered: %d", m.crawlFound),
				fmt.Sprintf("Visited: %d", m.crawlVisited),
			}
			if m.crawlCurrent != "" {
				lines = append(lines, fmt.Sprintf("Current: %s", m.crawlCurrent))
			}
			return loadingStyle.Render(strings.Join(lines, "\n"))
		}
		if m.currentDownloadURI == "" {
			return loadingStyle.Render(fmt.Sprintf("%s %s", m.spinner.View(), m.statusMsg))
		}
		current := m.downloadingDone + 1
		if m.downloadingDone >= m.downloading {
			current = m.downloadingDone
		}
		lines := []string{
			fmt.Sprintf("%s Downloading %d/%d", m.spinner.View(), current, m.downloading),
			fmt.Sprintf("Current: %s", m.currentDownloadURI),
			fmt.Sprintf("Output: %s", m.currentDownloadPath),
		}
		if m.lastDownloadStatus != "" {
			lines = append(lines, fmt.Sprintf("Last: %s", m.lastDownloadStatus))
		}
		return loadingStyle.Render(strings.Join(lines, "\n"))

	case StateError:
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit, backspace to go back", m.err))

	case StateBrowsing:
		// Build breadcrumb trail
		var breadcrumb string
		if len(m.breadcrumbs) > 0 {
			parts := make([]string, len(m.breadcrumbs))
			for i, bc := range m.breadcrumbs {
				parts[i] = bc.title
			}
			breadcrumb = breadcrumbStyle.Render(strings.Join(parts, " > "))
		}

		// Status bar with selection count or status message
		var status string
		if m.statusMsg != "" {
			status = statusStyle.Render(" " + m.statusMsg)
		} else if count := len(m.selected); count > 0 {
			status = statusStyle.Render(fmt.Sprintf(" %d selected • Press 'd' to download", count))
		}

		// Help
		help := helpStyle.Render("↑↓ nav • enter browse • space select • a all • d download • c clear • q quit")

		return docStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				breadcrumb,
				m.list.View(),
				status,
				help,
			),
		)

	default:
		return "Unknown state"
	}
}

// Styles
var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Padding(2, 4)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(2, 4)

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)
