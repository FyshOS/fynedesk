// +build linux openbsd freebsd netbsd

package wm

import (
	"github.com/BurntSushi/xgb/xproto"

	"fyne.io/fynedesk"
	"fyne.io/fynedesk/internal/x11"
)

type stack struct {
	clients      []fynedesk.Window
	mappingOrder []fynedesk.Window

	listeners []fynedesk.StackListener
}

func (s *stack) AddWindow(win fynedesk.Window) {
	if win == nil {
		return
	}
	s.addToStack(win)

	for _, l := range s.listeners {
		l.WindowAdded(win)
	}
}

func (s *stack) RaiseToTop(win fynedesk.Window) {
	if win.Iconic() {
		return
	}
	if len(s.clients) > 1 {
		win.RaiseAbove(s.TopWindow())
	}

	if s.indexForWin(win) == -1 {
		return
	}
	s.removeFromStack(win)
	s.addToStack(win)

	wm := fynedesk.Instance().WindowManager().(*x11WM)
	windowClientListStackingUpdate(wm)
}

func (s *stack) RemoveWindow(win fynedesk.Window) {
	s.removeFromStack(win)

	if s.TopWindow() != nil {
		s.TopWindow().Focus()
	}

	for _, l := range s.listeners {
		l.WindowRemoved(win)
	}
}

func (s *stack) TopWindow() fynedesk.Window {
	if len(s.clients) == 0 {
		return nil
	}
	return s.clients[len(s.clients)-1]
}

func (s *stack) Windows() []fynedesk.Window {
	return s.clients
}

func (s *stack) addToStack(win fynedesk.Window) {
	s.clients = append(s.clients, win)
	s.mappingOrder = append(s.mappingOrder, win.(x11.XWin))
}

func (s *stack) clientForWin(id xproto.Window) x11.XWin {
	for _, w := range s.clients {
		if w.(x11.XWin).FrameID() == id || w.(x11.XWin).ChildID() == id {
			return w.(x11.XWin)
		}
	}

	return nil
}

func (s *stack) getWindowsFromClients(clients []fynedesk.Window) []xproto.Window {
	var wins []xproto.Window
	for _, cli := range clients {
		wins = append(wins, cli.(x11.XWin).ChildID())
	}
	return wins
}

func (s *stack) indexForWin(win fynedesk.Window) int {
	pos := -1
	for i, w := range s.clients {
		if w == win {
			pos = i
		}
	}
	return pos
}

func (s *stack) removeFromStack(win fynedesk.Window) {
	pos := s.indexForWin(win)

	if pos == -1 {
		return
	}
	s.clients = append(s.clients[:pos], s.clients[pos+1:]...)

	pos = -1
	for i, w := range s.mappingOrder {
		if w == win {
			pos = i
		}
	}
	if pos == -1 {
		return
	}
	s.mappingOrder = append(s.mappingOrder[:pos], s.mappingOrder[pos+1:]...)
}
