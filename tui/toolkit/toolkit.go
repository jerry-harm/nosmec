package toolkit

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type keymapEntry struct {
	id    string
	help  string
	keys  []string
	bind  key.Binding
}

type ToolKit struct {
	mu sync.RWMutex

	focusedID string
	keymaps   map[string]keymapEntry
	viewcache map[string]string
	handlers  map[string]func(tea.Msg) tea.Cmd
}

func New() *ToolKit {
	return &ToolKit{
		keymaps:   make(map[string]keymapEntry),
		viewcache: make(map[string]string),
		handlers:  make(map[string]func(tea.Msg) tea.Cmd),
	}
}

// KeymapAdd registers a keymap entry with an ID, help text, and keys.
func (tk *ToolKit) KeymapAdd(id, help string, keys ...string) {
	tk.mu.Lock()
	defer tk.mu.Unlock()

	bind := key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(strings.Join(keys, ","), help),
	)

	tk.keymaps[id] = keymapEntry{
		id:   id,
		help: help,
		keys: keys,
		bind: bind,
	}
}

// KeymapGet retrieves a keymap entry by ID.
func (tk *ToolKit) KeymapGet(id string) (keymapEntry, bool) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()
	entry, ok := tk.keymaps[id]
	return entry, ok
}

// KeymapHelpStrings returns sorted help strings for all registered keymaps.
func (tk *ToolKit) KeymapHelpStrings() []string {
	tk.mu.RLock()
	defer tk.mu.RUnlock()

	var lines []string
	for _, entry := range tk.keymaps {
		line := fmt.Sprintf("%s: %s", strings.Join(entry.keys, "/"), entry.help)
		lines = append(lines, line)
	}
	sort.Strings(lines)
	return lines
}

// Focus sets the focused window ID.
func (tk *ToolKit) Focus(id string) {
	tk.mu.Lock()
	defer tk.mu.Unlock()
	tk.focusedID = id
}

// Blur clears the focused state.
func (tk *ToolKit) Blur() {
	tk.mu.Lock()
	defer tk.mu.Unlock()
	tk.focusedID = ""
}

// IsFocused returns true if the given ID is the currently focused window.
func (tk *ToolKit) IsFocused(id string) bool {
	tk.mu.RLock()
	defer tk.mu.RUnlock()
	return tk.focusedID == id
}

// CacheView stores a cached view string for a window.
func (tk *ToolKit) CacheView(id string, view string) {
	tk.mu.Lock()
	defer tk.mu.Unlock()
	tk.viewcache[id] = view
}

// GetCachedView retrieves a cached view string.
func (tk *ToolKit) GetCachedView(id string) (string, bool) {
	tk.mu.RLock()
	defer tk.mu.RUnlock()
	v, ok := tk.viewcache[id]
	return v, ok
}

// SetMsgHandling sets a message handler for a window ID.
func (tk *ToolKit) SetMsgHandling(id string, handler func(tea.Msg) tea.Cmd) {
	tk.mu.Lock()
	defer tk.mu.Unlock()
	tk.handlers[id] = handler
}

// HandleMsg calls the handler for a window ID if registered.
func (tk *ToolKit) HandleMsg(id string, msg tea.Msg) tea.Cmd {
	tk.mu.RLock()
	defer tk.mu.RUnlock()
	if handler, ok := tk.handlers[id]; ok {
		return handler(msg)
	}
	return nil
}

// FocusedID returns the currently focused window ID (read-only).
func (tk *ToolKit) FocusedID() string {
	tk.mu.RLock()
	defer tk.mu.RUnlock()
	return tk.focusedID
}
