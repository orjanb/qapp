package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/orjan/spotify/internal/lastfm"
	spotify "github.com/orjan/spotify/internal/spotify"
)

// ── views ────────────────────────────────────────────────────────────────────

type view int

const (
	viewSearch view = iota
	viewResults
	viewQueue
)

// ── list items ───────────────────────────────────────────────────────────────

type trackItem struct {
	track          spotify.Track
	prefix         string
	playCount      *int // nil = not yet loaded
	playCountFailed bool
}

func (t trackItem) Title() string {
	title := t.prefix + t.track.Name
	if t.playCountFailed {
		return fmt.Sprintf("%s  (no play data available)", title)
	}
	if t.playCount != nil {
		return fmt.Sprintf("%s  (%d plays)", title, *t.playCount)
	}
	return title
}

func (t trackItem) Description() string {
	return spotify.ArtistNames(t.track) + " · " + t.track.Album.Name
}

func (t trackItem) FilterValue() string { return t.track.Name }

// ── messages ─────────────────────────────────────────────────────────────────

type searchResultsMsg struct {
	tracks []spotify.Track
	err    error
}

type queueLoadedMsg struct {
	resp *spotify.QueueResponse
	err  error
}

type addedToQueueMsg struct {
	track spotify.Track
	err   error
}

type skippedToNextMsg struct {
	err error
}

type nowPlayingMsg struct {
	resp *spotify.PlaybackStateResponse
	err  error
}

type nowPlayingStatsMsg struct {
	stats *lastfm.TrackStats
	err   error
}

type queueItemStatsMsg struct {
	index   int
	trackID string
	stats   *lastfm.TrackStats
	err     error
}

type resultItemStatsMsg struct {
	index   int
	trackID string
	stats   *lastfm.TrackStats
	err     error
}

type tickMsg time.Time

// ── styling ──────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1DB954"))
	statusOK   = lipgloss.NewStyle().Foreground(lipgloss.Color("#1DB954"))
	statusErr  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	appStyle   = lipgloss.NewStyle().Padding(0, 1)
)

// ── model ────────────────────────────────────────────────────────────────────

type Model struct {
	client          *spotify.Client
	lastfm          *lastfm.Client
	ctx             context.Context
	currentView     view
	prevView        view
	input           textinput.Model
	resultsList     list.Model
	queueList       list.Model
	status          string
	statusIsErr     bool
	windowWidth     int
	windowHeight    int
	nowPlaying      *spotify.PlaybackStateResponse
	nowPlayingStats *lastfm.TrackStats
	progressBar     progress.Model
	progressMs      int
	tickCount       int
	lastTrackID     string
}

func New(client *spotify.Client, lfm *lastfm.Client, ctx context.Context) Model {
	ti := textinput.New()
	ti.Placeholder = "Search for a track…"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 40

	delegate := list.NewDefaultDelegate()

	rl := list.New([]list.Item{}, delegate, 0, 0)
	rl.Title = "Results"
	rl.SetShowStatusBar(false)
	rl.SetFilteringEnabled(false)
	rl.DisableQuitKeybindings()

	ql := list.New([]list.Item{}, delegate, 0, 0)
	ql.Title = "Queue"
	ql.SetShowStatusBar(false)
	ql.SetFilteringEnabled(false)
	ql.DisableQuitKeybindings()

	pb := progress.New(progress.WithDefaultGradient())

	return Model{
		client:      client,
		lastfm:      lfm,
		ctx:         ctx,
		currentView: viewSearch,
		input:       ti,
		resultsList: rl,
		queueList:   ql,
		progressBar: pb,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, doLoadNowPlaying(m.client, m.ctx), doTick())
}

// ── update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		overhead := 9
		if m.lastfm != nil {
			overhead = 10
		}
		listHeight := msg.Height - overhead
		m.resultsList.SetSize(msg.Width-2, listHeight)
		m.queueList.SetSize(msg.Width-2, listHeight)
		m.input.Width = msg.Width - 4
		m.progressBar.Width = (msg.Width - 4) / 4
		return m, nil

	case searchResultsMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusIsErr = true
			return m, nil
		}
		items := make([]list.Item, len(msg.tracks))
		for i, t := range msg.tracks {
			items[i] = trackItem{track: t}
		}
		m.resultsList.SetItems(items)
		m.currentView = viewResults
		m.status = fmt.Sprintf("%d results", len(msg.tracks))
		m.statusIsErr = false
		if m.lastfm != nil {
			var cmds []tea.Cmd
			for i, t := range msg.tracks {
				cmds = append(cmds, doLoadResultItemStats(m.lastfm, m.ctx, i, t))
			}
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case queueLoadedMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusIsErr = true
			return m, nil
		}
		var items []list.Item
		for _, t := range msg.resp.Queue {
			items = append(items, trackItem{track: t})
		}
		m.queueList.SetItems(items)
		m.status = fmt.Sprintf("%d items in queue", len(items))
		m.statusIsErr = false
		if m.lastfm != nil {
			var cmds []tea.Cmd
			for i, item := range items {
				ti := item.(trackItem)
				cmds = append(cmds, doLoadQueueItemStats(m.lastfm, m.ctx, i, ti.track))
			}
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case addedToQueueMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusIsErr = true
			return m, nil
		}
		m.status = fmt.Sprintf("Added %q to queue", msg.track.Name)
		m.statusIsErr = false
		return m, nil

	case skippedToNextMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusIsErr = true
			return m, nil
		}
		m.status = "Skipped"
		m.statusIsErr = false
		return m, doLoadNowPlaying(m.client, m.ctx)

	case tickMsg:
		m.tickCount++
		if m.nowPlaying != nil && m.nowPlaying.IsPlaying {
			m.progressMs += 1000
			d := m.nowPlaying.Item.DurationMs
			if d > 0 && m.progressMs >= d {
				return m, tea.Batch(doTick(), doLoadNowPlaying(m.client, m.ctx))
			}
		}
		var cmd tea.Cmd
		if m.tickCount%10 == 0 {
			cmd = tea.Batch(doTick(), doLoadNowPlaying(m.client, m.ctx))
		} else {
			cmd = doTick()
		}
		return m, cmd

	case nowPlayingMsg:
		if msg.err == nil {
			m.nowPlaying = msg.resp
			if msg.resp != nil {
				m.progressMs = msg.resp.ProgressMs
				if msg.resp.Item.ID != m.lastTrackID {
					m.lastTrackID = msg.resp.Item.ID
					m.nowPlayingStats = nil
					cmds := []tea.Cmd{doLoadQueue(m.client, m.ctx)}
					if m.lastfm != nil {
						cmds = append(cmds, doLoadNowPlayingStats(m.lastfm, m.ctx, *msg.resp.Item))
					}
					return m, tea.Batch(cmds...)
				}
			}
		}
		return m, nil

	case nowPlayingStatsMsg:
		if msg.err == nil {
			m.nowPlayingStats = msg.stats
		}
		return m, nil

	case queueItemStatsMsg:
		if msg.err != nil || msg.stats == nil {
			items := m.queueList.Items()
			if msg.index < len(items) {
				if ti, ok := items[msg.index].(trackItem); ok && ti.track.ID == msg.trackID {
					ti.playCountFailed = true
					m.queueList.SetItem(msg.index, ti)
				}
			}
			return m, nil
		}
		items := m.queueList.Items()
		if msg.index >= len(items) {
			return m, nil
		}
		ti, ok := items[msg.index].(trackItem)
		if !ok || ti.track.ID != msg.trackID {
			return m, nil
		}
		count := msg.stats.UserPlayCount
		ti.playCount = &count
		m.queueList.SetItem(msg.index, ti)
		return m, nil

	case resultItemStatsMsg:
		if msg.err != nil || msg.stats == nil {
			items := m.resultsList.Items()
			if msg.index < len(items) {
				if ti, ok := items[msg.index].(trackItem); ok && ti.track.ID == msg.trackID {
					ti.playCountFailed = true
					m.resultsList.SetItem(msg.index, ti)
				}
			}
			return m, nil
		}
		items := m.resultsList.Items()
		if msg.index >= len(items) {
			return m, nil
		}
		ti, ok := items[msg.index].(trackItem)
		if !ok || ti.track.ID != msg.trackID {
			return m, nil
		}
		count := msg.stats.UserPlayCount
		ti.playCount = &count
		m.resultsList.SetItem(msg.index, ti)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {

	case viewSearch:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			query := m.input.Value()
			if query == "" {
				return m, nil
			}
			m.status = "Searching…"
			m.statusIsErr = false
			return m, doSearch(m.client, m.ctx, query)
		case tea.KeyTab:
			m.prevView = viewSearch
			m.currentView = viewQueue
			m.status = "Loading queue…"
			m.statusIsErr = false
			return m, doLoadQueue(m.client, m.ctx)
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case viewResults:
		switch {
		case msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc:
			return m, tea.Quit
		case msg.Type == tea.KeyEnter:
			selected, ok := m.resultsList.SelectedItem().(trackItem)
			if !ok {
				return m, nil
			}
			m.status = fmt.Sprintf("Adding %q...", selected.track.Name)
			m.statusIsErr = false
			return m, doAddToQueue(m.client, m.ctx, selected.track)
		case msg.String() == "/":
			m.currentView = viewSearch
			m.input.Focus()
			return m, tea.Batch(textinput.Blink, doLoadNowPlaying(m.client, m.ctx))
		case msg.Type == tea.KeyTab:
			m.prevView = viewResults
			m.currentView = viewQueue
			m.status = "Loading queue…"
			m.statusIsErr = false
			return m, doLoadQueue(m.client, m.ctx)
		default:
			var cmd tea.Cmd
			m.resultsList, cmd = m.resultsList.Update(msg)
			return m, cmd
		}

	case viewQueue:
		switch {
		case msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc:
			return m, tea.Quit
		case msg.String() == "n":
			m.status = "Skipping…"
			m.statusIsErr = false
			return m, doSkipToNext(m.client, m.ctx)
		case msg.String() == "r":
			m.status = "Refreshing…"
			m.statusIsErr = false
			return m, doLoadQueue(m.client, m.ctx)
		case msg.Type == tea.KeyTab:
			m.currentView = m.prevView
			if m.prevView == viewSearch {
				m.input.Focus()
				return m, tea.Batch(textinput.Blink, doLoadNowPlaying(m.client, m.ctx))
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.queueList, cmd = m.queueList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// ── now playing ──────────────────────────────────────────────────────────────

func (m Model) nowPlayingView() string {
	if m.nowPlaying == nil {
		return helpStyle.Render("  Nothing playing")
	}
	symbol := "⏸"
	if m.nowPlaying.IsPlaying {
		symbol = "▶"
	}
	line1 := titleStyle.Render(symbol + "  " + m.nowPlaying.Item.Name)
	line2 := helpStyle.Render("   " + spotify.ArtistNames(*m.nowPlaying.Item) + " · " + m.nowPlaying.Item.Album.Name)

	var pct float64
	if d := m.nowPlaying.Item.DurationMs; d > 0 {
		pct = float64(m.progressMs) / float64(d)
		if pct > 1.0 {
			pct = 1.0
		}
	}
	bar := m.progressBar.ViewAs(pct)
	timeStr := helpStyle.Render(fmt.Sprintf(" %s / %s", formatDuration(m.progressMs), formatDuration(m.nowPlaying.Item.DurationMs)))

	line1line2 := lipgloss.JoinHorizontal(lipgloss.Center, line1, "  ", bar, timeStr) + "\n" + line2
	if m.nowPlayingStats != nil {
		line3 := helpStyle.Render(fmt.Sprintf(
			"   ♪ %d plays  •  %d listeners",
			m.nowPlayingStats.UserPlayCount,
			m.nowPlayingStats.Listeners,
		))
		return line1line2 + "\n" + line3
	}
	return line1line2
}

func formatDuration(ms int) string {
	s := ms / 1000
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// ── view ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	viewName := map[view]string{
		viewSearch:  "Search",
		viewResults: "Results",
		viewQueue:   "Queue",
	}[m.currentView]

	header := titleStyle.Render("Spotify") + "  ·  " + viewName
	var body string

	switch m.currentView {
	case viewSearch:
		body = "\n" + m.nowPlayingView() + "\n\n  " + m.input.View()
	case viewResults:
		body = "\n" + m.nowPlayingView() + "\n\n" + m.resultsList.View()
	case viewQueue:
		body = "\n" + m.nowPlayingView() + "\n\n" + m.queueList.View()
	}

	var statusLine string
	if m.status != "" {
		if m.statusIsErr {
			statusLine = statusErr.Render(m.status)
		} else {
			statusLine = statusOK.Render(m.status)
		}
	}

	help := helpStyle.Render(m.helpText())

	return appStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			body,
			statusLine,
			help,
		),
	)
}

func (m Model) helpText() string {
	switch m.currentView {
	case viewSearch:
		return "enter: search  •  tab: queue  •  esc/ctrl+c: quit"
	case viewResults:
		return "enter: add to queue  •  /: search  •  tab: queue  •  esc/ctrl+c: quit"
	case viewQueue:
		return "n: skip  •  r: refresh  •  tab: back  •  esc/ctrl+c: quit"
	}
	return ""
}

// ── commands ──────────────────────────────────────────────────────────────────

func doSearch(client *spotify.Client, ctx context.Context, query string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := client.SearchTracks(ctx, query, 10)
		return searchResultsMsg{tracks: tracks, err: err}
	}
}

func doLoadQueue(client *spotify.Client, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetQueue(ctx)
		return queueLoadedMsg{resp: resp, err: err}
	}
}

func doAddToQueue(client *spotify.Client, ctx context.Context, track spotify.Track) tea.Cmd {
	return func() tea.Msg {
		err := client.AddToQueue(ctx, track.URI)
		return addedToQueueMsg{track: track, err: err}
	}
}

func doSkipToNext(client *spotify.Client, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		err := client.SkipToNext(ctx)
		return skippedToNextMsg{err: err}
	}
}

func doLoadNowPlaying(client *spotify.Client, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetPlaybackState(ctx)
		return nowPlayingMsg{resp: resp, err: err}
	}
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func doLoadNowPlayingStats(lfm *lastfm.Client, ctx context.Context, track spotify.Track) tea.Cmd {
	return func() tea.Msg {
		if len(track.Artists) == 0 {
			return nowPlayingStatsMsg{err: fmt.Errorf("no artists")}
		}
		stats, err := lfm.GetTrackInfo(ctx, track.Artists[0].Name, track.Name)
		return nowPlayingStatsMsg{stats: stats, err: err}
	}
}

func doLoadQueueItemStats(lfm *lastfm.Client, ctx context.Context, index int, track spotify.Track) tea.Cmd {
	return func() tea.Msg {
		if len(track.Artists) == 0 {
			return queueItemStatsMsg{index: index, trackID: track.ID, err: fmt.Errorf("no artists")}
		}
		stats, err := lfm.GetTrackInfo(ctx, track.Artists[0].Name, track.Name)
		return queueItemStatsMsg{index: index, trackID: track.ID, stats: stats, err: err}
	}
}

func doLoadResultItemStats(lfm *lastfm.Client, ctx context.Context, index int, track spotify.Track) tea.Cmd {
	return func() tea.Msg {
		if len(track.Artists) == 0 {
			return resultItemStatsMsg{index: index, trackID: track.ID, err: fmt.Errorf("no artists")}
		}
		stats, err := lfm.GetTrackInfo(ctx, track.Artists[0].Name, track.Name)
		return resultItemStatsMsg{index: index, trackID: track.ID, stats: stats, err: err}
	}
}
