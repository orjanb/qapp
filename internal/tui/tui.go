package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	track  spotify.Track
	prefix string
}

func (t trackItem) Title() string {
	return t.prefix + t.track.Name
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
	client       *spotify.Client
	ctx          context.Context
	currentView  view
	prevView     view
	input        textinput.Model
	resultsList  list.Model
	queueList    list.Model
	status       string
	statusIsErr  bool
	windowWidth  int
	windowHeight int
}

func New(client *spotify.Client, ctx context.Context) Model {
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

	return Model{
		client:      client,
		ctx:         ctx,
		currentView: viewSearch,
		input:       ti,
		resultsList: rl,
		queueList:   ql,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// ── update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		listHeight := msg.Height - 5
		m.resultsList.SetSize(msg.Width-2, listHeight)
		m.queueList.SetSize(msg.Width-2, listHeight)
		m.input.Width = msg.Width - 4
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
		return m, nil

	case queueLoadedMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			m.statusIsErr = true
			return m, nil
		}
		var items []list.Item
		if msg.resp.CurrentlyPlaying != nil {
			items = append(items, trackItem{
				track:  *msg.resp.CurrentlyPlaying,
				prefix: "▶ ",
			})
		}
		for _, t := range msg.resp.Queue {
			items = append(items, trackItem{track: t})
		}
		m.queueList.SetItems(items)
		m.status = fmt.Sprintf("%d items in queue", len(items))
		m.statusIsErr = false
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

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {

	case viewSearch:
		switch msg.Type {
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
		case msg.String() == "q":
			return m, tea.Quit
		case msg.Type == tea.KeyEnter:
			selected, ok := m.resultsList.SelectedItem().(trackItem)
			if !ok {
				return m, nil
			}
			m.status = fmt.Sprintf("Adding %q...", selected.track.Name)
			m.statusIsErr = false
			return m, doAddToQueue(m.client, m.ctx, selected.track)
		case msg.String() == "/" || msg.Type == tea.KeyEsc:
			m.currentView = viewSearch
			m.input.Focus()
			return m, textinput.Blink
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
		case msg.String() == "q":
			return m, tea.Quit
		case msg.String() == "r":
			m.status = "Refreshing…"
			m.statusIsErr = false
			return m, doLoadQueue(m.client, m.ctx)
		case msg.Type == tea.KeyTab || msg.Type == tea.KeyEsc:
			m.currentView = m.prevView
			if m.prevView == viewSearch {
				m.input.Focus()
				return m, textinput.Blink
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
		body = "\n  " + m.input.View()
	case viewResults:
		body = m.resultsList.View()
	case viewQueue:
		body = m.queueList.View()
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
		return "enter: search  •  tab: queue"
	case viewResults:
		return "enter: add to queue  •  /: search  •  tab: queue  •  q: quit"
	case viewQueue:
		return "r: refresh  •  tab/esc: back  •  q: quit"
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
