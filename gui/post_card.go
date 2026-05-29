package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func buildPostActionRow(onOpenThread func()) fyne.CanvasObject {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.VisibilityIcon(), func() {
			if onOpenThread != nil {
				onOpenThread()
			}
		}),
		widget.NewToolbarAction(theme.MailReplyIcon(), func() {}),
		widget.NewToolbarAction(theme.MoreHorizontalIcon(), func() {}),
	)
}

type threadEntryCard struct {
	widget.BaseWidget

	card  *widget.Card
	onTap func()
}

func newThreadEntryCard(content fyne.CanvasObject, onTap func()) *threadEntryCard {
	entry := &threadEntryCard{
		card:  widget.NewCard("", "", content),
		onTap: onTap,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (c *threadEntryCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.card)
}

func (c *threadEntryCard) Tapped(*fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

func (c *threadEntryCard) TappedSecondary(*fyne.PointEvent) {}

type accordionStateBridge struct {
	widget.BaseWidget

	accordion *widget.Accordion
	onOpen    func()
	onClose   func()
	wasOpen   bool
}

func newAccordionStateBridge(
	accordion *widget.Accordion,
	onOpen func(),
	onClose func(),
) fyne.CanvasObject {
	bridge := &accordionStateBridge{
		accordion: accordion,
		onOpen:    onOpen,
		onClose:   onClose,
		wasOpen:   len(accordion.Items) > 0 && accordion.Items[0].Open,
	}
	bridge.ExtendBaseWidget(bridge)
	return bridge
}

func (a *accordionStateBridge) CreateRenderer() fyne.WidgetRenderer {
	return &accordionStateBridgeRenderer{bridge: a, renderer: widget.NewSimpleRenderer(a.accordion)}
}

type accordionStateBridgeRenderer struct {
	bridge   *accordionStateBridge
	renderer fyne.WidgetRenderer
}

func (r *accordionStateBridgeRenderer) Destroy() {
	r.renderer.Destroy()
}

func (r *accordionStateBridgeRenderer) Layout(size fyne.Size) {
	r.renderer.Layout(size)
}

func (r *accordionStateBridgeRenderer) MinSize() fyne.Size {
	return r.renderer.MinSize()
}

func (r *accordionStateBridgeRenderer) Objects() []fyne.CanvasObject {
	return r.renderer.Objects()
}

func (r *accordionStateBridgeRenderer) Refresh() {
	isOpen := len(r.bridge.accordion.Items) > 0 && r.bridge.accordion.Items[0].Open
	if isOpen != r.bridge.wasOpen {
		if isOpen {
			if r.bridge.onOpen != nil {
				r.bridge.onOpen()
			}
		} else if r.bridge.onClose != nil {
			r.bridge.onClose()
		}
		r.bridge.wasOpen = isOpen
	}

	r.renderer.Refresh()
}

type tightVBoxLayout struct {
	spacing float32
}

func newTightVBoxLayout(spacing float32) fyne.Layout {
	return &tightVBoxLayout{spacing: spacing}
}

func buildReplyCard(replies []Reply, extraCount int, preview bool, onOpenThread func()) fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(replies)+2)
	rows = append(rows, newStyledLabel(T("replies"), fyne.TextStyle{Bold: true}))

	for _, reply := range replies {
		body := reply.Body
		if preview {
			body = truncateBody(body, 72)
		}

		row := widget.NewLabel(truncateNpub(reply.Author) + ": " + body)
		row.Wrapping = fyne.TextWrapWord
		rows = append(rows, row)
	}

	if preview && extraCount > 0 {
		rows = append(rows, newStyledLabel(
			fmt.Sprintf(T("more_replies"), extraCount),
			fyne.TextStyle{Italic: true},
		))
	}

	compactContent := container.New(
		newTightVBoxLayout(theme.InnerPadding()/4),
		rows...,
	)

	background := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	background.StrokeColor = theme.Color(theme.ColorNameSeparator)
	background.StrokeWidth = 1
	inset := container.New(
		layout.NewCustomPaddedLayout(
			theme.InnerPadding()/2,
			theme.InnerPadding()/2,
			theme.InnerPadding()/2,
			theme.InnerPadding()/2,
		),
		compactContent,
	)
	surface := container.NewStack(background, inset)
	content := container.New(
		layout.NewCustomPaddedLayout(
			theme.InnerPadding()/4,
			theme.InnerPadding()/4,
			theme.InnerPadding()/4,
			theme.InnerPadding()/4,
		),
		surface,
	)
	if onOpenThread == nil {
		return content
	}

	return newThreadEntryCard(content, onOpenThread)
}

func (l *tightVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	visibleCount := 0

	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}

		minSize := obj.MinSize()
		obj.Move(fyne.NewPos(0, y))
		obj.Resize(fyne.NewSize(size.Width, minSize.Height))
		y += minSize.Height
		visibleCount++
		if visibleCount < visibleObjects(objects) {
			y += l.spacing
		}
	}
}

func (l *tightVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width float32
	var height float32
	count := 0

	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}

		minSize := obj.MinSize()
		if minSize.Width > width {
			width = minSize.Width
		}
		height += minSize.Height
		count++
	}

	if count > 1 {
		height += float32(count-1) * l.spacing
	}

	return fyne.NewSize(width, height)
}

func visibleObjects(objects []fyne.CanvasObject) int {
	count := 0
	for _, obj := range objects {
		if obj.Visible() {
			count++
		}
	}
	return count
}
