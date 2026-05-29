package gui

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestTopBarHasModeTabsSearchAndActions(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()

	barContainer, modeTabs, searchEntry, actionButtons := buildTopBarLayout()
	if len(barContainer.Objects) != 3 {
		t.Fatalf("top bar region count = %d, want 3", len(barContainer.Objects))
	}

	if !containsObject(barContainer.Objects, modeTabs) {
		t.Fatal("top bar is missing left mode tab region")
	}

	if !containsObject(barContainer.Objects, searchEntry) {
		t.Fatal("top bar is missing center search region")
	}

	if !containsObject(barContainer.Objects, actionButtons) {
		t.Fatal("top bar is missing right action region")
	}

	assertButtonTexts(t, modeTabs, []string{T("community"), T("dm"), T("note")})

	if got := searchEntry.PlaceHolder; got != T("search") {
		t.Fatalf("search placeholder = %q, want %q", got, T("search"))
	}

	assertIconButtons(t, actionButtons, 2)
	if len(actionButtons.Objects) != 2 {
		t.Fatalf("action button count = %d, want 2", len(actionButtons.Objects))
	}
	if len(modeTabs.Objects) != 3 {
		t.Fatalf("mode tab count = %d, want 3", len(modeTabs.Objects))
	}
	_ = test.WidgetRenderer(searchEntry)
}

func TestPostCardPreservesExplicitToolbarActions(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()
	currentView = viewList
	currentPost = nil

	post := Post{
		ID:         "1",
		Author:     "npub1xyzxyzxyz",
		Community:  "tech",
		Body:       "hello world from a post card",
		ReplyCount: 1,
		Timestamp:  time.Now(),
		TopReplies: []Reply{{Author: "npub1reply", Body: "nested reply"}},
	}

	obj := buildPostCard(&post)
	if _, ok := obj.(fyne.Tappable); ok {
		t.Fatal("post card should not rely on whole-card tap behavior")
	}

	card, ok := obj.(*widget.Card)
	if !ok {
		t.Fatalf("post card type = %T, want *widget.Card", obj)
	}

	if !containsObjectType(card.Content, (*widget.Toolbar)(nil)) {
		t.Fatal("post card is missing explicit toolbar actions")
	}

	toolbar := firstToolbar(card.Content)
	if toolbar == nil {
		t.Fatal("expected toolbar actions")
	}

	if got := len(toolbar.Items); got == 0 {
		t.Fatal("toolbar has no actions")
	}

	if got := countToolbarActions(toolbar); got == 0 {
		t.Fatal("toolbar exposes no actionable items")
	}

	openAction, ok := toolbar.Items[0].(*widget.ToolbarAction)
	if !ok {
		t.Fatalf("first toolbar item type = %T, want *widget.ToolbarAction", toolbar.Items[0])
	}
	if openAction.OnActivated == nil {
		t.Fatal("open thread toolbar action should be wired")
	}

	openAction.OnActivated()
	if currentView != viewThread {
		t.Fatalf("toolbar open action did not switch view, got %v want %v", currentView, viewThread)
	}
	if currentPost == nil || currentPost.ID != post.ID {
		t.Fatalf("toolbar open action selected post = %#v, want post ID %q", currentPost, post.ID)
	}
	_ = test.WidgetRenderer(card)
}

func TestReplyCardIsPrimaryThreadEntryPoint(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()
	currentView = viewList
	currentPost = nil

	post := Post{
		ID:         "thread-post",
		Author:     "npub1xyzxyzxyz",
		Community:  "tech",
		Body:       "hello world from a post card",
		ReplyCount: 4,
		Timestamp:  time.Now(),
		TopReplies: []Reply{
			{Author: "npub1reply1", Body: "nested reply one"},
			{Author: "npub1reply2", Body: "nested reply two"},
		},
		ExtraCount: 2,
	}

	obj := buildPostCard(&post)
	card, ok := obj.(*widget.Card)
	if !ok {
		t.Fatalf("post card type = %T, want *widget.Card", obj)
	}

	replyCard := firstThreadEntryCard(card.Content)
	if replyCard == nil {
		t.Fatal("expected nested reply card thread entry")
	}

	replyCard.Tapped(&fyne.PointEvent{})
	if currentView != viewThread {
		t.Fatalf("reply card tap did not switch view, got %v want %v", currentView, viewThread)
	}
	if currentPost == nil || currentPost.ID != post.ID {
		t.Fatalf("reply card tap selected post = %#v, want post ID %q", currentPost, post.ID)
	}

	_ = test.WidgetRenderer(replyCard)
}

func TestPostCardRendersSingleNestedReplyCard(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()

	post := Post{
		ID:         "reply-card-post",
		Author:     "npub1xyzxyzxyz",
		Community:  "tech",
		Body:       "hello world from a post card",
		ReplyCount: 4,
		Timestamp:  time.Now(),
		TopReplies: []Reply{
			{Author: "npub1reply1", Body: "nested reply one"},
			{Author: "npub1reply2", Body: "nested reply two"},
		},
		ExtraCount: 2,
	}

	obj := buildPostCard(&post)
	card, ok := obj.(*widget.Card)
	if !ok {
		t.Fatalf("post card type = %T, want *widget.Card", obj)
	}

	if got := countThreadEntryCards(card.Content); got != 1 {
		t.Fatalf("nested reply card count = %d, want 1", got)
	}

	replyCard := firstThreadEntryCard(card.Content)
	if replyCard == nil {
		t.Fatal("expected nested reply card")
	}

	for _, want := range []string{
		T("replies"),
		"npub1reply1: nested reply one",
		"npub1reply2: nested reply two",
		fmt.Sprintf(T("more_replies"), post.ExtraCount),
	} {
		if !containsLabelText(replyCard.card.Content, want) {
			t.Fatalf("reply card missing label text %q", want)
		}
	}

	if !containsObjectType(card.Content, (*widget.Toolbar)(nil)) {
		t.Fatal("outer post action row should still be present")
	}
	_ = test.WidgetRenderer(card)
}

func TestReplyCardUsesNestedSurfaceStyling(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()

	replyCard, ok := buildReplyCard(
		[]Reply{{Author: "npub1reply1", Body: "nested reply one"}},
		0,
		true,
		func() {},
	).(*threadEntryCard)
	if !ok {
		t.Fatalf("reply card type = %T, want *threadEntryCard", replyCard)
	}

	background := firstRectangle(replyCard.card.Content)
	if background == nil {
		t.Fatal("reply card is missing nested surface background")
	}

	if got, want := background.FillColor, theme.Color(theme.ColorNameInputBackground); got != want {
		t.Fatalf("reply card fill color = %#v, want %#v", got, want)
	}

	if got, want := background.StrokeColor, theme.Color(theme.ColorNameSeparator); got != want {
		t.Fatalf("reply card stroke color = %#v, want %#v", got, want)
	}

	if background.StrokeWidth <= 0 {
		t.Fatalf("reply card stroke width = %v, want > 0", background.StrokeWidth)
	}

	_ = test.WidgetRenderer(replyCard)
}

func TestFullPostCardKeepsToolbarAndOmitsPreviewThreadEntry(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()

	post := Post{
		ID:         "post1",
		Author:     "npub1xyzxyzxyz",
		Community:  "tech",
		Body:       "full thread body",
		ReplyCount: 3,
		Timestamp:  time.Now(),
		TopReplies: []Reply{{Author: "npub1reply1", Body: "preview reply"}},
	}

	obj := buildPostCardFull(&post)
	card, ok := obj.(*widget.Card)
	if !ok {
		t.Fatalf("full post card type = %T, want *widget.Card", obj)
	}

	toolbar := firstToolbar(card.Content)
	if toolbar == nil {
		t.Fatal("full post card should keep outer post action toolbar")
	}

	if got := countToolbarActions(toolbar); got == 0 {
		t.Fatal("full post card toolbar exposes no actionable items")
	}

	if entry := firstThreadEntryCard(card.Content); entry != nil {
		t.Fatal("full post card should not include tappable nested reply entry")
	}

	if containsLabelText(card.Content, T("replies")) {
		t.Fatal("full post card should not render preview nested reply card")
	}

	_ = test.WidgetRenderer(card)
}

func TestSidebarShowsFeedGlobalAndCommunitySection(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()
	selectedScope = scopeMyFeed
	currentCommunity = ""
	currentView = viewList
	currentPost = nil
	communitiesCollapsed = false
	communitiesList = []string{"nostr-lang-zh", "music"}
	updateUI = func() {}

	sidebar := buildSidebar()
	scroll, ok := sidebar.(*container.Scroll)
	if !ok {
		t.Fatalf("sidebar type = %T, want *container.Scroll", sidebar)
	}

	content, ok := scroll.Content.(*fyne.Container)
	if !ok {
		t.Fatalf("sidebar content type = %T, want *fyne.Container", scroll.Content)
	}

	if got := len(content.Objects); got != 3 {
		t.Fatalf("sidebar section count = %d, want 3", got)
	}

	assertButtonTextAt(t, content.Objects[0], T("my_feed"))
	assertButtonTextAt(t, content.Objects[1], T("global"))

	bridge, ok := content.Objects[2].(*accordionStateBridge)
	if !ok {
		t.Fatalf("community section type = %T, want *accordionStateBridge", content.Objects[2])
	}

	if got := len(bridge.accordion.Items); got != 1 {
		t.Fatalf("community accordion item count = %d, want 1", got)
	}

	item := bridge.accordion.Items[0]
	if item.Title != T("communities") {
		t.Fatalf("community section title = %q, want %q", item.Title, T("communities"))
	}

	detail, ok := item.Detail.(*fyne.Container)
	if !ok {
		t.Fatalf("community section detail type = %T, want *fyne.Container", item.Detail)
	}

	assertButtonTexts(t, detail, []string{"nostr-lang-zh", "music"})
	_ = test.WidgetRenderer(bridge)
}

func TestSidebarCommunityButtonTapUpdatesSelection(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()
	selectedScope = scopeMyFeed
	currentCommunity = ""
	currentView = viewThread
	currentPost = &Post{ID: "stale"}
	communitiesCollapsed = false
	communitiesList = []string{"nostr-lang-zh", "music"}
	updateUI = func() {}

	sidebar := buildSidebar()
	bridge := firstAccordionBridge(sidebar)
	if bridge == nil {
		t.Fatal("expected community accordion bridge")
	}

	detail, ok := bridge.accordion.Items[0].Detail.(*fyne.Container)
	if !ok {
		t.Fatalf("community section detail type = %T, want *fyne.Container", bridge.accordion.Items[0].Detail)
	}

	musicButton, ok := detail.Objects[1].(*widget.Button)
	if !ok {
		t.Fatalf("community button type = %T, want *widget.Button", detail.Objects[1])
	}

	test.Tap(musicButton)

	if selectedScope != scopeCommunity {
		t.Fatalf("selectedScope = %v, want %v", selectedScope, scopeCommunity)
	}
	if currentCommunity != "music" {
		t.Fatalf("currentCommunity = %q, want %q", currentCommunity, "music")
	}
	if currentView != viewList {
		t.Fatalf("currentView = %v, want %v", currentView, viewList)
	}
	if currentPost != nil {
		t.Fatalf("currentPost = %#v, want nil", currentPost)
	}
}

func TestCommunitySectionRefreshTracksCollapsedState(t *testing.T) {
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("LC_ALL", "en_US.UTF-8")
	t.Setenv("FYNE_LOCALE", "en")
	SetLocale("en")

	test.NewApp()
	communitiesCollapsed = false

	section := buildCommunitySection([]fyne.CanvasObject{widget.NewLabel("nostr-lang-zh")})
	bridge, ok := section.(*accordionStateBridge)
	if !ok {
		t.Fatalf("community section type = %T, want *accordionStateBridge", section)
	}

	bridge.accordion.CloseAll()
	bridge.Refresh()
	if !communitiesCollapsed {
		t.Fatal("communitiesCollapsed = false, want true after close")
	}

	bridge.accordion.Open(0)
	bridge.Refresh()
	if communitiesCollapsed {
		t.Fatal("communitiesCollapsed = true, want false after reopen")
	}
	_ = test.WidgetRenderer(bridge)
}

func containsObject(objects []fyne.CanvasObject, target fyne.CanvasObject) bool {
	for _, object := range objects {
		if object == target {
			return true
		}
	}

	return false
}

func containsObjectType(root fyne.CanvasObject, want any) bool {
	if root == nil {
		return false
	}

	switch target := want.(type) {
	case *widget.Toolbar:
		if _, ok := root.(*widget.Toolbar); ok {
			return true
		}
		_ = target
	}

	if card, ok := root.(*widget.Card); ok {
		return containsObjectType(card.Content, want)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if containsObjectType(child, want) {
				return true
			}
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		return containsObjectType(s.Content, want)
	}

	return false
}

func firstToolbar(root fyne.CanvasObject) *widget.Toolbar {
	if root == nil {
		return nil
	}

	if toolbar, ok := root.(*widget.Toolbar); ok {
		return toolbar
	}

	if card, ok := root.(*widget.Card); ok {
		return firstToolbar(card.Content)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if toolbar := firstToolbar(child); toolbar != nil {
				return toolbar
			}
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		return firstToolbar(s.Content)
	}

	return nil
}

func firstAccordionBridge(root fyne.CanvasObject) *accordionStateBridge {
	if root == nil {
		return nil
	}

	if bridge, ok := root.(*accordionStateBridge); ok {
		return bridge
	}

	if card, ok := root.(*widget.Card); ok {
		return firstAccordionBridge(card.Content)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if bridge := firstAccordionBridge(child); bridge != nil {
				return bridge
			}
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		return firstAccordionBridge(s.Content)
	}

	return nil
}

func firstThreadEntryCard(root fyne.CanvasObject) *threadEntryCard {
	if root == nil {
		return nil
	}

	if entry, ok := root.(*threadEntryCard); ok {
		return entry
	}

	if card, ok := root.(*widget.Card); ok {
		return firstThreadEntryCard(card.Content)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if entry := firstThreadEntryCard(child); entry != nil {
				return entry
			}
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		return firstThreadEntryCard(s.Content)
	}

	return nil
}

func countToolbarActions(toolbar *widget.Toolbar) int {
	count := 0
	for _, item := range toolbar.Items {
		if _, ok := item.(*widget.ToolbarAction); ok {
			count++
		}
	}

	return count
}

func countThreadEntryCards(root fyne.CanvasObject) int {
	if root == nil {
		return 0
	}

	count := 0
	if _, ok := root.(*threadEntryCard); ok {
		count++
	}

	if card, ok := root.(*widget.Card); ok {
		count += countThreadEntryCards(card.Content)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			count += countThreadEntryCards(child)
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		count += countThreadEntryCards(s.Content)
	}

	return count
}

func firstRectangle(root fyne.CanvasObject) *canvas.Rectangle {
	if root == nil {
		return nil
	}

	if rect, ok := root.(*canvas.Rectangle); ok {
		return rect
	}

	if entry, ok := root.(*threadEntryCard); ok {
		return firstRectangle(entry.card.Content)
	}

	if card, ok := root.(*widget.Card); ok {
		return firstRectangle(card.Content)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if rect := firstRectangle(child); rect != nil {
				return rect
			}
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		return firstRectangle(s.Content)
	}

	return nil
}

func containsLabelText(root fyne.CanvasObject, want string) bool {
	if root == nil {
		return false
	}

	if label, ok := root.(*widget.Label); ok && label.Text == want {
		return true
	}

	if card, ok := root.(*widget.Card); ok {
		return containsLabelText(card.Content, want)
	}

	if c, ok := root.(*fyne.Container); ok {
		for _, child := range c.Objects {
			if containsLabelText(child, want) {
				return true
			}
		}
	}

	if s, ok := root.(*container.Scroll); ok {
		return containsLabelText(s.Content, want)
	}

	return false
}

func assertButtonTexts(t *testing.T, group *fyne.Container, want []string) {
	t.Helper()

	if len(group.Objects) != len(want) {
		t.Fatalf("button count = %d, want %d", len(group.Objects), len(want))
	}

	for i, expected := range want {
		button, ok := group.Objects[i].(*widget.Button)
		if !ok {
			t.Fatalf("button %d type = %T, want *widget.Button", i, group.Objects[i])
		}

		if button.Text != expected {
			t.Fatalf("button %d text = %q, want %q", i, button.Text, expected)
		}
	}
}

func assertIconButtons(t *testing.T, group *fyne.Container, want int) {
	t.Helper()

	if len(group.Objects) != want {
		t.Fatalf("icon button count = %d, want %d", len(group.Objects), want)
	}

	for i, object := range group.Objects {
		button, ok := object.(*widget.Button)
		if !ok {
			t.Fatalf("action button %d type = %T, want *widget.Button", i, object)
		}

		if button.Icon == nil {
			t.Fatalf("action button %d icon = nil, want non-nil", i)
		}
	}
}

func assertButtonTextAt(t *testing.T, object fyne.CanvasObject, want string) {
	t.Helper()

	button, ok := object.(*widget.Button)
	if !ok {
		t.Fatalf("button type = %T, want *widget.Button", object)
	}

	if button.Text != want {
		t.Fatalf("button text = %q, want %q", button.Text, want)
	}
}
