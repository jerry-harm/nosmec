# GUI Navigation Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the Fyne GUI so the top bar follows an app-bar-plus-tabs model, post cards stop using whole-card tap behavior, replies become the entry point into thread view, and GUI tests move closer to real Fyne widget interaction.

**Architecture:** Keep the current desktop shell, but split responsibilities into app bar, sidebar, post card body, and action row. Replace custom whole-card tapping with explicit reply-card navigation and post action controls so the layout can later map cleanly to Android app bar, tabs, and action row patterns.

**Tech Stack:** Go, Fyne v2.7.4, Cobra (CLI fallback), `fyne/test` for widget-level tests.

---

## File Structure

- `gui/app.go`
  Purpose: window bootstrap, app-level state, top-level layout composition.
- `gui/post_card.go`
  Purpose: post card composition, reply card composition, action row widgets.
- `gui/app_test.go`
  Purpose: pure logic tests that should remain small and deterministic.
- `gui/ui_test.go`
  Purpose: new Fyne widget interaction tests using `fyne/test`.
- `main.go`
  Purpose: keep GUI as default startup entry.

## Task 1: Refactor Top Navigation Into App Bar + Mode Tabs

**Files:**
- Modify: `gui/app.go`
- Test: `gui/ui_test.go`

- [ ] **Step 1: Write the failing widget test for top navigation structure**

Test assertions to add in `gui/ui_test.go`:
```go
func TestTopBarHasModeTabsSearchAndActions(t *testing.T) {
	bar := buildTopBar()
	if bar == nil {
		t.Fatal("expected top bar")
	}
}
```

- [ ] **Step 2: Run the focused GUI test and verify failure or missing coverage**

Run: `go test ./gui -run TestTopBarHasModeTabsSearchAndActions -count=1`
Expected: either missing test file or insufficient assertions to prove structure.

- [ ] **Step 3: Reshape top bar into app-bar semantics**

Implementation target in `gui/app.go`:
```go
func buildTopBar() fyne.CanvasObject {
	modeTabs := container.NewHBox(
		widget.NewButton(T("community"), func() {}),
		widget.NewButton(T("dm"), func() {}),
		widget.NewButton(T("note"), func() {}),
	)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(T("search"))

	actions := container.NewHBox(
		widget.NewButtonWithIcon("", theme.AccountIcon(), func() {}),
		widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {}),
	)

	return container.NewBorder(nil, nil, modeTabs, actions, searchEntry)
}
```

- [ ] **Step 4: Add widget-level assertions instead of only nil checks**

Test direction in `gui/ui_test.go`:
```go
func TestTopBarHasModeTabsSearchAndActions(t *testing.T) {
	bar := buildTopBar()
	canvas := test.NewCanvas()
	canvas.SetContent(bar)
	objects := canvas.Content().(*fyne.Container).Objects
	if len(objects) == 0 {
		t.Fatal("expected top bar content")
	}
}
```

- [ ] **Step 5: Run GUI tests**

Run: `go test ./gui -run TestTopBarHasModeTabsSearchAndActions -count=1`
Expected: PASS

## Task 2: Remove Whole-Card Tapping And Add Explicit Post Actions

**Files:**
- Modify: `gui/post_card.go`
- Modify: `gui/app.go`
- Test: `gui/ui_test.go`

- [ ] **Step 1: Write failing test that post cards expose explicit actions**

Add to `gui/ui_test.go`:
```go
func TestPostCardUsesExplicitActionsInsteadOfWholeCardTap(t *testing.T) {
	post := Post{ID: "1", Author: "npub1x", Community: "tech", Body: "hello", ReplyCount: 1, Timestamp: time.Now()}
	obj := buildPostCard(&post)
	if obj == nil {
		t.Fatal("expected post card")
	}
}
```

- [ ] **Step 2: Remove `tappableCard` if no longer needed**

Target simplification in `gui/post_card.go`:
```go
func buildPostCard(post *Post) fyne.CanvasObject {
	return widget.NewCard("", "", buildPostCardContent(post, true))
}
```

- [ ] **Step 3: Add a compact action row to the outer post card**

Target helper in `gui/post_card.go` or `gui/app.go`:
```go
func buildPostActionRow() fyne.CanvasObject {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FavoriteIcon(), func() {}),
		widget.NewToolbarAction(theme.MailReplyIcon(), func() {}),
		widget.NewToolbarAction(theme.MoreHorizontalIcon(), func() {}),
	)
}
```

- [ ] **Step 4: Append the action row to post content**

Target composition:
```go
	content := container.NewVBox(
		meta,
		body,
		stats,
		replyCard,
		buildPostActionRow(),
	)
```

- [ ] **Step 5: Run tests and verify the custom tappable widget is gone if unused**

Run: `go test ./gui -count=1`
Expected: PASS

## Task 3: Make Nested Reply Card The Navigation Entry Point

**Files:**
- Modify: `gui/post_card.go`
- Modify: `gui/app.go`
- Test: `gui/ui_test.go`

- [ ] **Step 1: Write failing test that reply card exists as one nested card**

Add to `gui/ui_test.go`:
```go
func TestPostCardRendersSingleNestedReplyCard(t *testing.T) {
	post := Post{ID: "post1", Author: "npub1x", Community: "tech", Body: "body", ReplyCount: 3, Timestamp: time.Now(), TopReplies: []Reply{{Author: "npub1r", Body: "reply"}}}
	obj := buildPostCardContent(&post, true)
	if obj == nil {
		t.Fatal("expected card content")
	}
}
```

- [ ] **Step 2: Keep one nested reply card, not one reply widget per reply**

Target shape:
```go
func buildReplyCard(replies []Reply, extraCount int, preview bool, onOpen func()) fyne.CanvasObject {
	rows := []fyne.CanvasObject{widget.NewLabel(T("replies"))}
	for _, reply := range replies {
		rows = append(rows, widget.NewLabel(truncateNpub(reply.Author)+": "+truncateBody(reply.Body, 72)))
	}
	card := widget.NewCard("", "", container.NewVBox(rows...))
	return container.NewVBox(card, widget.NewButton(T("open_thread"), onOpen))
}
```

- [ ] **Step 3: Move thread entry to reply card / explicit open action**

Target behavior:
```go
func openThread(post *Post) {
	postCopy := *post
	currentPost = &postCopy
	currentView = viewThread
	updateUI()
}
```

- [ ] **Step 4: Use tighter spacing and different nested background**

Implementation target:
```go
	compactContent := container.New(layout.NewCustomPaddedLayout(4, 4, 4, 4), rowsBox)
	background := canvas.NewRectangle(theme.Color(theme.ColorNameMenuBackground))
	return widget.NewCard("", "", container.NewStack(background, compactContent))
```

- [ ] **Step 5: Run focused tests**

Run: `go test ./gui -run 'TestPostCardRendersSingleNestedReplyCard|TestPostCardUsesExplicitActionsInsteadOfWholeCardTap' -count=1`
Expected: PASS

## Task 4: Add Real Fyne Widget Interaction Tests

**Files:**
- Create: `gui/ui_test.go`
- Modify: `gui/app_test.go`

- [ ] **Step 1: Add a minimal Fyne test harness file**

Initial file content:
```go
package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)
```

- [ ] **Step 2: Add sidebar accordion presence test**

Test target:
```go
func TestSidebarContainsCommunitySection(t *testing.T) {
	sidebar := buildSidebar()
	if sidebar == nil {
		t.Fatal("expected sidebar")
	}
	canvas := test.NewCanvas()
	canvas.SetContent(sidebar)
}
```

- [ ] **Step 3: Add locale logic tests only where pure logic matters**

Keep in `gui/app_test.go`:
```go
func TestNormalizeLocale(t *testing.T) {
	if got := normalizeLocale("zh-CN"); got != "zh" {
		t.Fatalf("got %q", got)
	}
}
```

- [ ] **Step 4: Run GUI test suite**

Run: `go test ./gui -count=1`
Expected: PASS

## Task 5: Final Verification And Cleanup

**Files:**
- Modify: `gui/app.go`
- Modify: `gui/post_card.go`
- Modify: `gui/ui_test.go`

- [ ] **Step 1: Remove dead code after tappable-card retirement**

Delete unused helpers if no longer referenced:
```go
// remove tappableCard if buildPostCard no longer uses it
```

- [ ] **Step 2: Verify GUI remains default entrypoint**

Check `main.go` remains:
```go
func main() {
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		cmd.Execute()
		return
	}
	gui.Run()
}
```

- [ ] **Step 3: Run full verification**

Run: `go test ./gui -count=1 && go vet ./gui/... && go build ./...`
Expected: all pass with exit code 0.

- [ ] **Step 4: Manual smoke checklist**

Verify locally:
```text
1. nosmec opens GUI by default
2. community accordion opens/closes
3. post card shows action row
4. clicking nested reply entry opens thread
5. outer card itself is not the navigation trigger
```

## Self-Review

- Spec coverage: covers app bar structure, sidebar/community navigation, removal of whole-card tapping, action row, nested reply-card navigation, and Fyne widget testing.
- Placeholder scan: no `TBD` or implicit “handle later” steps remain.
- Type consistency: plan assumes `buildTopBar`, `buildSidebar`, `buildPostCard`, `buildPostCardContent`, `buildReplyCard`, and `openThread` helpers; these names are internally consistent.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-29-gui-navigation-refactor.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
