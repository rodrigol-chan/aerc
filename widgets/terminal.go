package widgets

import (
	"os/exec"
	"syscall"

	"git.sr.ht/~rjarry/aerc/lib/ui"
	"git.sr.ht/~rjarry/aerc/log"
	tcellterm "git.sr.ht/~rockorager/tcell-term"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type Terminal struct {
	closed    bool
	cmd       *exec.Cmd
	ctx       *ui.Context
	destroyed bool
	focus     bool
	vterm     *tcellterm.Terminal
	running   bool

	OnClose func(err error)
	OnEvent func(event tcell.Event) bool
	OnStart func()
	OnTitle func(title string)
}

func NewTerminal(cmd *exec.Cmd) (*Terminal, error) {
	term := &Terminal{
		cmd:   cmd,
		vterm: tcellterm.New(),
	}
	return term, nil
}

func (term *Terminal) Close(err error) {
	if term.closed {
		return
	}
	if term.vterm != nil {
		// Stop receiving events
		term.vterm.Unwatch(term)
		term.vterm.Close()
	}
	if !term.closed && term.OnClose != nil {
		term.OnClose(err)
	}
	if term.ctx != nil {
		term.ctx.HideCursor()
	}
	term.closed = true
}

func (term *Terminal) Destroy() {
	if term.destroyed {
		return
	}
	if term.ctx != nil {
		term.ctx.HideCursor()
	}
	// If we destroy, we don't want to call the OnClose callback
	term.OnClose = nil
	term.Close(nil)
	term.vterm = nil
	term.destroyed = true
}

func (term *Terminal) Invalidate() {
	ui.Invalidate()
}

func (term *Terminal) Draw(ctx *ui.Context) {
	if term.destroyed {
		return
	}
	term.ctx = ctx // gross
	term.vterm.SetView(ctx.View())
	if !term.running && !term.closed && term.cmd != nil {
		term.vterm.Watch(term)
		attr := &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 1}
		if err := term.vterm.StartWithAttrs(term.cmd, attr); err != nil {
			log.Errorf("error running terminal: %v", err)
			term.Close(err)
			return
		}
		term.running = true
		if term.OnStart != nil {
			term.OnStart()
		}
	}
	term.draw()
}

func (term *Terminal) draw() {
	term.vterm.Draw()
	if term.focus && !term.closed && term.ctx != nil {
		vis, x, y, style := term.vterm.GetCursor()
		if vis {
			term.ctx.SetCursor(x, y)
			term.ctx.SetCursorStyle(style)
		} else {
			term.ctx.HideCursor()
		}
	}
}

func (term *Terminal) MouseEvent(localX int, localY int, event tcell.Event) {
	ev, ok := event.(*tcell.EventMouse)
	if !ok {
		return
	}
	if term.OnEvent != nil {
		term.OnEvent(ev)
	}
	if term.closed {
		return
	}
	e := tcell.NewEventMouse(localX, localY, ev.Buttons(), ev.Modifiers())
	term.vterm.HandleEvent(e)
}

func (term *Terminal) Focus(focus bool) {
	if term.closed {
		return
	}
	term.focus = focus
	if term.ctx != nil {
		if !term.focus {
			term.ctx.HideCursor()
		} else {
			_, x, y, style := term.vterm.GetCursor()
			term.ctx.SetCursor(x, y)
			term.ctx.SetCursorStyle(style)
			term.Invalidate()
		}
	}
}

// HandleEvent is used to watch the underlying terminal events
func (term *Terminal) HandleEvent(ev tcell.Event) bool {
	if term.closed || term.destroyed {
		return false
	}
	switch ev := ev.(type) {
	case *views.EventWidgetContent:
		ui.QueueRedraw()
		return true
	case *tcellterm.EventTitle:
		if term.OnTitle != nil {
			term.OnTitle(ev.Title())
		}
	case *tcellterm.EventClosed:
		term.Close(nil)
		ui.QueueRedraw()
	}
	return false
}

func (term *Terminal) Event(event tcell.Event) bool {
	if term.OnEvent != nil {
		if term.OnEvent(event) {
			return true
		}
	}
	if term.closed {
		return false
	}
	return term.vterm.HandleEvent(event)
}
