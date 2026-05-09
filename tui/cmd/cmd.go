package cmd

// Window management messages

type WinOpen struct {
	ID     string
	Window any // the window/model to open
}

type WinClose struct {
	ID string
}

type WinFocus struct {
	ID string
}

type WinBlur struct {
	ID string
}

type ViewFocus struct {
	WinID string
}

type ViewBlur struct {
	WinID string
}

type WinFreshData struct {
	ID string
}

type WinRefreshData struct {
	ID string
}

type MsgError struct {
	Err error
}
