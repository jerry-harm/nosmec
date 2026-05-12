package windowmanager

import (
	"sync"

	tea "charm.land/bubbletea/v2"
	"fiatjaf.com/nostr"
	"github.com/jerry-harm/nosmec/config"
	"github.com/jerry-harm/nosmec/logger"
	"github.com/jerry-harm/nosmec/tui/compose"
	"github.com/jerry-harm/nosmec/tui/window"
)

const ComposeWindowID = "compose"

type WindowManager struct {
	mu           sync.RWMutex
	windows      map[string]window.Window
	stack        []string // ordered by z-order (first = bottom, last = top)
	focused      string
	app          *config.AppContext
	composeModel window.Window
}

func New() *WindowManager {
	return &WindowManager{
		windows: make(map[string]window.Window),
		stack:   make([]string, 0),
	}
}

func (wm *WindowManager) SetApp(app *config.AppContext) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.app = app
}

func (wm *WindowManager) ComposeModel() window.Window {
	if wm.composeModel == nil && wm.app != nil {
		wm.composeModel = compose.NewModel(wm.app)
	}
	return wm.composeModel
}

func (wm *WindowManager) PrepareReply(event *nostr.Event) {
	compose.PrepareReply(wm.ComposeModel(), event)
}

func (wm *WindowManager) PrepareQuote(event *nostr.Event) {
	compose.PrepareQuote(wm.ComposeModel(), event)
}

func (wm *WindowManager) OpenCompose() tea.Cmd {
	if wm.composeModel == nil && wm.app != nil {
		wm.composeModel = compose.NewModel(wm.app)
	}
	id := wm.composeModel.ID()
	wm.windows[id] = wm.composeModel
	wm.removeFromStack(id)
	wm.stack = append(wm.stack, id)
	wm.focused = id
	return wm.composeModel.Init()
}

func (wm *WindowManager) openLocked(win window.Window) tea.Cmd {
	id := win.ID()
	wm.windows[id] = win
	wm.removeFromStack(id)
	wm.stack = append(wm.stack, id)
	wm.focused = id
	return win.Init()
}

// Open adds a window to the manager and focuses it. It calls Init() on the window
// and returns the resulting tea.Cmd.
func (wm *WindowManager) Open(win window.Window) tea.Cmd {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	id := win.ID()
	wm.windows[id] = win

	// Move to top of stack
	wm.removeFromStack(id)
	wm.stack = append(wm.stack, id)
	wm.focused = id

	// Call Init on the window and return its command
	return win.Init()
}

// Close removes a window from the manager.
func (wm *WindowManager) Close(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	logger.Debug("WindowManager.Close called", "id", id)
	delete(wm.windows, id)
	wm.removeFromStack(id)

	if wm.focused == id {
		// Focus next in stack
		if len(wm.stack) > 0 {
			wm.focused = wm.stack[len(wm.stack)-1]
		} else {
			wm.focused = ""
		}
	}
}

// Focus brings a window to the front and sets it as focused.
func (wm *WindowManager) Focus(id string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, ok := wm.windows[id]; !ok {
		return
	}

	wm.removeFromStack(id)
	wm.stack = append(wm.stack, id)
	wm.focused = id
}

// IsOpen returns true if the window is open.
func (wm *WindowManager) IsOpen(id string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	_, ok := wm.windows[id]
	return ok
}

// IsFocused returns true if the window is the focused one.
func (wm *WindowManager) IsFocused(id string) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.focused == id
}

// Update sends a message to all windows.
func (wm *WindowManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var cmd tea.Cmd
	for id, win := range wm.windows {
		_, c := win.Update(msg)
		if c != nil {
			cmd = c
		}
		_ = id
	}

	// Return the topmost window as the model for tea.WindowSizeMsg propagation
	if wm.focused != "" {
		if topWin, ok := wm.windows[wm.focused]; ok {
			return topWin, cmd
		}
	}
	return nil, cmd
}

// UpdateAll sends a message to all windows and returns all commands.
func (wm *WindowManager) UpdateAll(msg tea.Msg) ([]tea.Model, []tea.Cmd) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	var models []tea.Model
	var cmds []tea.Cmd

	for _, id := range wm.stack {
		win := wm.windows[id]
		m, c := win.Update(msg)
		models = append(models, m)
		if c != nil {
			cmds = append(cmds, c)
		}
		_ = id
	}

	return models, cmds
}

// UpdateFocused sends a message only to the focused window.
func (wm *WindowManager) UpdateFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if wm.focused == "" {
		return nil, nil
	}

	win, ok := wm.windows[wm.focused]
	if !ok {
		return nil, nil
	}

	return win.Update(msg)
}

// Resize sends a window size message to all windows.
func (wm *WindowManager) Resize(width, height int) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	msg := tea.WindowSizeMsg{Width: width, Height: height}
	for _, win := range wm.windows {
		win.Update(msg)
	}
}

// ResizeAll sends a window size message to all windows.
func (wm *WindowManager) ResizeAll(width, height int) {
	wm.Resize(width, height)
}

// View renders only the topmost window's full view.
// The topmost window's tea.View with AltScreen takes over the full screen.
func (wm *WindowManager) View() string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if len(wm.stack) == 0 {
		return ""
	}
	topID := wm.stack[len(wm.stack)-1]
	return wm.windows[topID].View().Content
}

// Get returns a window by ID.
func (wm *WindowManager) Get(id string) (window.Window, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	win, ok := wm.windows[id]
	return win, ok
}

// FocusedWindow returns the currently focused window.
func (wm *WindowManager) FocusedWindow() (window.Window, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if wm.focused == "" {
		return nil, false
	}
	win, ok := wm.windows[wm.focused]
	return win, ok
}

// WindowCount returns the number of open windows.
func (wm *WindowManager) WindowCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.windows)
}

// removeFromStack removes an ID from the stack.
func (wm *WindowManager) removeFromStack(id string) {
	for i, v := range wm.stack {
		if v == id {
			wm.stack = append(wm.stack[:i], wm.stack[i+1:]...)
			return
		}
	}
}
