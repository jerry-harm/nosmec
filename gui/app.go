package gui

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jerry-harm/nosmec/config"
)

var currentLocale = "en"

var translations = map[string]map[string]string{
	"en": {
		"community":       "Community",
		"dm":              "DM",
		"note":            "Note",
		"search":          "Search...",
		"my_feed":         "My Feed",
		"global":          "Global",
		"communities":     "Communities",
		"no_posts":        "No posts yet",
		"no_communities":  "No communities",
		"replies":         "Replies",
		"more_replies":    "%d more replies...",
		"back":            "← Back",
		"author":          "Author:",
		"community_label": "Community:",
		"just_now":        "just now",
		"ago_min":         "%dm ago",
		"ago_hour":        "%dh ago",
		"ago_day":         "%dd ago",
	},
	"zh": {
		"community":       "社区",
		"dm":              "私信",
		"note":            "笔记",
		"search":          "搜索...",
		"my_feed":         "我的动态",
		"global":          "全局",
		"communities":     "社区列表",
		"no_posts":        "暂无帖子",
		"no_communities":  "暂无社区",
		"replies":         "回复",
		"more_replies":    "还有 %d 条回复...",
		"back":            "← 返回",
		"author":          "作者：",
		"community_label": "社区：",
		"just_now":        "刚刚",
		"ago_min":         "%d分钟前",
		"ago_hour":        "%d小时前",
		"ago_day":         "%d天前",
	},
}

func T(key string) string {
	if locale, ok := translations[currentLocale]; ok {
		if s, ok := locale[key]; ok {
			return s
		}
	}
	return key
}

func SetLocale(locale string) {
	currentLocale = normalizeLocale(locale)
}

type viewMode int

const (
	viewList viewMode = iota
	viewThread
)

type scope int

const (
	scopeMyFeed scope = iota
	scopeGlobal
	scopeCommunity
)

var (
	defaultCommunitiesList = []string{"nostr-lang-zh", "music", "tech"}

	currentView      viewMode = viewList
	selectedScope    scope    = scopeMyFeed
	currentPost      *Post
	currentCommunity string

	communitiesCollapsed = false
	communitiesList      = append([]string(nil), defaultCommunitiesList...)

	mockPosts   []Post
	mockReplies map[string][]Reply
)

type Post struct {
	ID         string
	Author     string
	Community  string
	Body       string
	ReplyCount int
	Timestamp  time.Time
	TopReplies []Reply
	ExtraCount int
}

type Reply struct {
	ID        string
	Author    string
	Body      string
	Timestamp time.Time
}

func init() {
	now := time.Now()
	postTime := func(hoursAgo int) time.Time {
		return now.Add(-time.Duration(hoursAgo) * time.Hour)
	}

	mockPosts = []Post{
		{
			ID:         "post1",
			Author:     "npub1abc123def456ghi789jkl012mno345pqr678stu901vwx234yz",
			Community:  "nostr-lang-zh",
			Body:       "This is the first post in the nostr-lang-zh community. Really excited to see how Fyne is shaping up for desktop apps!",
			ReplyCount: 5,
			Timestamp:  postTime(2),
			ExtraCount: 2,
		},
		{
			ID:         "post2",
			Author:     "npub1xyz987wvu654tsr321qpo098nml765kji654gfe321dcb210",
			Community:  "music",
			Body:       "Just discovered an amazing album that blend jazz and electronic music. Has anyone else heard of this genre fusion?",
			ReplyCount: 3,
			Timestamp:  postTime(5),
			ExtraCount: 0,
		},
		{
			ID:         "post3",
			Author:     "npub1aaa111bbb222ccc333dddeee444fff555666ggg777hhh888iii",
			Community:  "tech",
			Body:       "What are everyone's thoughts on the latest developments in decentralized social networks? I think the future is promising.",
			ReplyCount: 12,
			Timestamp:  postTime(8),
			ExtraCount: 9,
		},
		{
			ID:         "post4",
			Author:     "npub1nnn888mmm777lll666kkk555jjj444iii333hhh222ggg111fff000",
			Community:  "nostr-lang-zh",
			Body:       "分享一个新项目：基于 Nostr 的去中心化社交平台实验。代码还在早期阶段，欢迎围观！",
			ReplyCount: 7,
			Timestamp:  postTime(12),
			ExtraCount: 4,
		},
		{
			ID:         "post5",
			Author:     "npub1ppp999qqq888rrr777sss666ttt555uuu444vvv333www222xxx111",
			Community:  "music",
			Body:       "Looking for recommendations on ambient electronic albums that are good for deep work sessions. Something atmospheric and minimal.",
			ReplyCount: 2,
			Timestamp:  postTime(24),
			ExtraCount: 0,
		},
		{
			ID:         "post6",
			Author:     "npub1kkk000jjj111iii222hhh333ggg444fff555eee666ddd777ccc888",
			Community:  "tech",
			Body:       "Has anyone tried the new Fyne v2.7 features for custom themes? I'd love to hear about custom widget styling approaches.",
			ReplyCount: 8,
			Timestamp:  postTime(36),
			ExtraCount: 5,
		},
	}

	mockReplies = map[string][]Reply{
		"post1": {
			{ID: "r1a", Author: "npub1rep1...", Body: "Great post! Fyne is really nice to work with.", Timestamp: postTime(1)},
			{ID: "r1b", Author: "npub1rep2...", Body: "I agree, the declarative style is clean.", Timestamp: postTime(1)},
			{ID: "r1c", Author: "npub1rep3...", Body: "Any examples of desktop widgets?", Timestamp: postTime(0)},
		},
		"post2": {
			{ID: "r2a", Author: "npub1rep4...", Body: "Check out some of the newer jazz-electronic releases!", Timestamp: postTime(4)},
			{ID: "r2b", Author: "npub1rep5...", Body: "Genre fusion is amazing right now.", Timestamp: postTime(3)},
			{ID: "r2c", Author: "npub1rep6...", Body: "Would love a playlist link!", Timestamp: postTime(2)},
		},
		"post3": {
			{ID: "r3a", Author: "npub1rep7...", Body: "Decentralized is the way to go for social.", Timestamp: postTime(7)},
			{ID: "r3b", Author: "npub1rep8...", Body: "Privacy first, always.", Timestamp: postTime(6)},
			{ID: "r3c", Author: "npub1rep9...", Body: "Nostr has great potential.", Timestamp: postTime(5)},
		},
		"post4": {
			{ID: "r4a", Author: "npub1rep0...", Body: "看起来很有意思，支持一下！", Timestamp: postTime(11)},
			{ID: "r4b", Author: "npub1repA...", Body: "代码仓库在哪里？", Timestamp: postTime(10)},
			{ID: "r4c", Author: "npub1repB...", Body: "期待正式版发布", Timestamp: postTime(9)},
		},
		"post5": {
			{ID: "r5a", Author: "npub1repC...", Body: "Try Boards of Canada,他们的作品很适合专注。", Timestamp: postTime(20)},
			{ID: "r5b", Author: "npub1repD...", Body: "Aphex Twin的ambient作品也不错。", Timestamp: postTime(18)},
			{ID: "r5c", Author: "npub1repE...", Body: "Brian Eno是经典之选。", Timestamp: postTime(16)},
		},
		"post6": {
			{ID: "r6a", Author: "npub1repF...", Body: "Custom themes in Fyne are quite flexible!", Timestamp: postTime(30)},
			{ID: "r6b", Author: "npub1repG...", Body: "Check the了他们最新的theme示例。", Timestamp: postTime(28)},
			{ID: "r6c", Author: "npub1repH...", Body: "我也在研究Fyne主题定制。", Timestamp: postTime(24)},
		},
	}

	for i := range mockPosts {
		post := &mockPosts[i]
		replies := mockReplies[post.ID]
		if len(replies) == 0 {
			continue
		}

		n := min(3, len(replies))
		post.TopReplies = replies[:n]
	}
}

func Run() {
	SetLocale(string(lang.SystemLocale()))
	if os.Getenv("FYNE_LOCALE") == "" {
		_ = os.Setenv("FYNE_LOCALE", currentLocale)
	}

	communitiesList = loadCommunitiesList()

	a := app.New()
	w := a.NewWindow("nosmec")
	w.Resize(fyne.NewSize(1024, 768))

	contentArea := container.NewStack()
	sidebarArea := container.NewStack()

	updateUI = func() {
		sidebarArea.Objects = []fyne.CanvasObject{buildSidebar()}
		sidebarArea.Refresh()

		contentArea.Objects = []fyne.CanvasObject{buildContent()}
		contentArea.Refresh()
	}

	topBar := buildTopBar()
	split := container.NewHSplit(sidebarArea, contentArea)
	split.SetOffset(0.25)

	mainContent := container.NewBorder(topBar, nil, nil, nil, split)
	minBounds := canvas.NewRectangle(color.Transparent)
	minBounds.SetMinSize(fyne.NewSize(800, 600))
	w.SetContent(container.NewStack(minBounds, mainContent))
	updateUI()
	w.ShowAndRun()
}

var updateUI func()

func buildTopBar() fyne.CanvasObject {
	bar, _, _, _ := buildTopBarLayout()
	return bar
}

func buildTopBarLayout() (*fyne.Container, *fyne.Container, *widget.Entry, *fyne.Container) {
	modeTabs, searchEntry, actionButtons := buildTopBarRegions()

	bar := container.NewBorder(nil, nil, modeTabs, actionButtons, searchEntry)

	return bar, modeTabs, searchEntry, actionButtons
}

func buildTopBarRegions() (*fyne.Container, *widget.Entry, *fyne.Container) {
	communityBtn := widget.NewButton(T("community"), func() {})
	communityBtn.Importance = widget.HighImportance

	dmBtn := widget.NewButton(T("dm"), func() {})
	noteBtn := widget.NewButton(T("note"), func() {})
	modeTabs := container.NewHBox(communityBtn, dmBtn, noteBtn)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(T("search"))

	profileBtn := widget.NewButtonWithIcon("", theme.AccountIcon(), func() {})
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {})
	actionButtons := container.NewHBox(profileBtn, settingsBtn)

	return modeTabs, searchEntry, actionButtons
}

func buildSidebar() fyne.CanvasObject {
	feedBtn := newSidebarButton(T("my_feed"), selectedScope == scopeMyFeed, func() {
		selectedScope = scopeMyFeed
		currentView = viewList
		currentPost = nil
		updateUI()
	})

	globalBtn := newSidebarButton(T("global"), selectedScope == scopeGlobal, func() {
		selectedScope = scopeGlobal
		currentView = viewList
		currentPost = nil
		updateUI()
	})

	communityButtons := buildCommunityButtons()
	communitySection := buildCommunitySection(communityButtons)

	sidebarContent := container.NewVBox(feedBtn, globalBtn, communitySection)
	return container.NewScroll(sidebarContent)
}

func buildCommunityButtons() []fyne.CanvasObject {
	if len(communitiesList) == 0 {
		return []fyne.CanvasObject{widget.NewLabel(T("no_communities"))}
	}

	buttons := make([]fyne.CanvasObject, 0, len(communitiesList))
	for _, name := range communitiesList {
		communityName := name
		buttons = append(buttons, newSidebarButton(
			communityName,
			selectedScope == scopeCommunity && currentCommunity == communityName,
			func() {
				selectedScope = scopeCommunity
				currentCommunity = communityName
				currentView = viewList
				currentPost = nil
				updateUI()
			},
		))
	}

	return buttons
}

func buildCommunitySection(buttons []fyne.CanvasObject) fyne.CanvasObject {
	detail := container.NewVBox(buttons...)
	item := widget.NewAccordionItem(T("communities"), detail)
	item.Open = !communitiesCollapsed

	accordion := widget.NewAccordion(item)
	accordion.MultiOpen = false

	return newAccordionStateBridge(
		accordion,
		func() {
			communitiesCollapsed = false
		},
		func() {
			communitiesCollapsed = true
		},
	)
}

func buildContent() fyne.CanvasObject {
	if currentView == viewThread && currentPost != nil {
		return buildThreadView()
	}
	return buildListView()
}

func buildListView() fyne.CanvasObject {
	posts := postsForScope(mockPosts, selectedScope, currentCommunity)

	if len(posts) == 0 {
		emptyLabel := widget.NewLabel(T("no_posts"))
		emptyLabel.Alignment = fyne.TextAlignCenter
		return container.NewCenter(emptyLabel)
	}

	cards := make([]fyne.CanvasObject, 0, len(posts))
	for _, post := range posts {
		postCopy := post
		cards = append(cards, buildPostCard(&postCopy))
	}

	return container.NewScroll(container.NewVBox(cards...))
}

func buildThreadView() fyne.CanvasObject {
	if currentPost == nil {
		return widget.NewLabel("No post selected")
	}

	backBtn := widget.NewButton(T("back"), func() {
		currentView = viewList
		updateUI()
	})

	threadContainer := container.NewVBox(
		backBtn,
		buildPostCardFull(currentPost),
	)

	return container.NewScroll(threadContainer)
}

func buildPostCard(post *Post) fyne.CanvasObject {
	return widget.NewCard("", "", buildPostCardContent(post, true))
}

func buildPostCardFull(post *Post) fyne.CanvasObject {
	return widget.NewCard("", "", buildPostCardContent(post, false))
}

func buildPostCardContent(post *Post, preview bool) fyne.CanvasObject {
	body := post.Body
	if preview {
		body = truncateBody(body, 100)
	}

	bodyLabel := widget.NewLabel(body)
	bodyLabel.Wrapping = fyne.TextWrapWord

	statsLabel := widget.NewLabel(
		fmt.Sprintf("%d %s  •  %s", post.ReplyCount, T("replies"), formatTimestamp(post.Timestamp)),
	)

	content := container.NewVBox(
		newStyledLabel(T("author")+" "+truncateNpub(post.Author), fyne.TextStyle{Bold: true}),
		newStyledLabel(T("community_label")+" "+post.Community, fyne.TextStyle{Italic: true}),
		bodyLabel,
		statsLabel,
	)

	replies := post.TopReplies
	if !preview {
		replies = mockReplies[post.ID]
	}

	if len(replies) > 0 {
		var onOpen func()
		if preview {
			onOpen = func() {
				openThread(post)
			}
		}

		content.Add(buildReplyCard(replies, post.ExtraCount, preview, onOpen))
	}

	if preview {
		content.Add(buildPostActionRow(func() {
			openThread(post)
		}))
	}

	return content
}

func openThread(post *Post) {
	if post == nil {
		return
	}

	postCopy := *post
	currentPost = &postCopy
	currentView = viewThread
	if updateUI != nil {
		updateUI()
	}
}

func newStyledLabel(text string, style fyne.TextStyle) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = style
	label.Wrapping = fyne.TextWrapWord
	return label
}

func newSidebarButton(label string, selected bool, onTap func()) *widget.Button {
	btn := widget.NewButton(label, onTap)
	if selected {
		btn.Importance = widget.HighImportance
	}
	return btn
}

func loadCommunitiesList() []string {
	cfg := config.InitConfig()
	return communityNamesFromSubscriptions(cfg.Subscriptions, defaultCommunitiesList)
}

func communityNamesFromSubscriptions(
	subscriptions []config.Subscription,
	fallback []string,
) []string {
	seen := map[string]struct{}{}
	communities := make([]string, 0, len(subscriptions))

	for _, sub := range subscriptions {
		if sub.Type != "community" {
			continue
		}

		key := strings.TrimSpace(sub.ID)
		name := strings.TrimSpace(sub.Petname)
		if name == "" {
			name = key
		}
		if name == "" {
			continue
		}
		if key == "" {
			key = name
		}
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		communities = append(communities, name)
	}

	if len(communities) > 0 {
		return communities
	}

	return append([]string(nil), fallback...)
}

func postsForScope(posts []Post, selected scope, community string) []Post {
	if selected != scopeCommunity {
		return append([]Post(nil), posts...)
	}

	filtered := make([]Post, 0, len(posts))
	for _, post := range posts {
		if post.Community == community {
			filtered = append(filtered, post)
		}
	}

	return filtered
}

func normalizeLocale(locale string) string {
	trimmed := strings.TrimSpace(locale)
	if trimmed == "" {
		return "en"
	}

	normalized := strings.ToLower(trimmed)
	if idx := strings.IndexAny(normalized, "-_"); idx >= 0 {
		normalized = normalized[:idx]
	}

	if _, ok := translations[normalized]; ok {
		return normalized
	}

	return "en"
}

func truncateNpub(np string) string {
	if len(np) <= 12 {
		return np
	}
	return np[:12] + "..."
}

func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "..."
}

func formatTimestamp(t time.Time) string {
	ago := time.Since(t)
	if ago < time.Minute {
		return T("just_now")
	}
	if ago < time.Hour {
		return fmt.Sprintf(T("ago_min"), int(ago.Minutes()))
	}
	if ago < 24*time.Hour {
		return fmt.Sprintf(T("ago_hour"), int(ago.Hours()))
	}
	return fmt.Sprintf(T("ago_day"), int(ago.Hours()/24))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
