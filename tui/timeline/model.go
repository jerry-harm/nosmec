package timeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/utils"
)

const maxItems = 100

type eventKind int

const (
	kindNote eventKind = iota
	kindReply
	kindQuote
	kindRepost
	kindCommunity
)

type eventItem struct {
	event      utils.TimelineEvent
	authorName string
	kind       eventKind
	width      int
}

func (i eventItem) Title() string {
	var prefix string
	if i.authorName != "" {
		prefix = "@" + i.authorName
	} else {
		prefix = "@" + i.event.Event.PubKey.Hex()[:8]
	}

	switch i.kind {
	case kindReply:
		prefix += " [Reply]"
	case kindQuote:
		prefix += " [Quote]"
	case kindRepost:
		prefix += " [Repost]"
	case kindCommunity:
		prefix += " [Community]"
	}

	return prefix
}

func (i eventItem) Description() string {
	content := i.event.Event.Content
	availableWidth := i.width - 10
	if availableWidth < 20 {
		availableWidth = 60
	}
	wrapped := lipgloss.Wrap(content, availableWidth, " \t\n")
	lines := strings.Split(wrapped, "\n")
	if len(lines) > 3 {
		lines = lines[:3]
	}
	return strings.Join(lines, "\n")
}

func (i eventItem) FilterValue() string {
	return i.event.Event.Content
}

func detectEventKind(event utils.TimelineEvent) eventKind {
	ev := event.Event
	if ev.Kind == 6 || ev.Kind == 16 {
		return kindRepost
	}

	if ev.Kind == 1111 {
		return kindCommunity
	}

	if ev.Kind == 1 {
		for _, tag := range ev.Tags {
			if len(tag) >= 2 {
				if tag[0] == "e" && len(tag) >= 4 && tag[3] == "reply" {
					return kindReply
				}
				if tag[0] == "q" {
					return kindQuote
				}
			}
		}
	}

	return kindNote
}

type TimelineModel struct {
	list     list.Model
	spinner  spinner.Model
	loading  bool
	err      error
	filter   string
	hashtags []string
	limit    int
	width    int
	height   int
	items    []eventItem
	app      *config.AppContext
}

type fetchMsg struct {
	events []utils.TimelineEvent
}

type errorMsg struct {
	err error
}

type usernameMsg struct {
	pubkey string
	name   string
}

func NewTimelineModel(app *config.AppContext, filter string, hashtags []string, limit int) *TimelineModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	width := 80
	height := 4

	items := make([]list.Item, 0)
	delegate := newTimelineDelegate(width, height)
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.SetFilteringEnabled(false)

	return &TimelineModel{
		list:     l,
		spinner:  s,
		loading:  true,
		filter:   filter,
		hashtags: hashtags,
		limit:    limit,
		width:    width,
		height:   height,
		app:      app,
	}
}

func (m *TimelineModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchTimeline(),
	)
}

func (m *TimelineModel) fetchTimeline() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if m.app == nil {
			return errorMsg{err: fmt.Errorf("app not initialized")}
		}

		opts := &utils.GetOptions{
			App: m.app,
		}

		var events []utils.TimelineEvent
		var err error

		switch m.filter {
		case "global":
			nostrEvents, e := utils.GetGlobalTimeline(ctx, m.limit, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		case "mine":
			nostrEvents, e := utils.GetMyTimeline(ctx, m.limit, opts)
			if e != nil {
				err = e
			} else {
				for _, e := range nostrEvents {
					events = append(events, utils.TimelineEvent{Event: e})
				}
			}
		default:
			events, err = utils.GetFollowedTimeline(ctx, m.limit, m.hashtags, opts)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		if len(events) > maxItems {
			events = events[:maxItems]
		}

		return fetchMsg{events: events}
	}
}

func (m *TimelineModel) refresh() {
	m.loading = true
	m.err = nil
}

func (m *TimelineModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

func (m *TimelineModel) IsLoading() bool {
	return m.loading
}

func (m *TimelineModel) GetError() error {
	return m.err
}

func (m *TimelineModel) GetFilter() string {
	return m.filter
}

func (m *TimelineModel) GetLimit() int {
	return m.limit
}

func (m *TimelineModel) GetHashtags() []string {
	return m.hashtags
}

func (m *TimelineModel) fetchUsername(pubkey string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var pk nostr.PubKey
		copy(pk[:], []byte(pubkey))
		name := utils.GetProfileName(ctx, pk, nil)
		return usernameMsg{pubkey: pubkey, name: name}
	}
}
