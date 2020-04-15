package wm

import (
	"fyne.io/fynedesk"
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
	return s.clients[0]
}

func (s *stack) Windows() []fynedesk.Window {
	return s.clients
}

func (s *stack) addToStack(win fynedesk.Window) {
	s.clients = append([]fynedesk.Window{win}, s.clients...)
	s.mappingOrder = append(s.mappingOrder, win)
}

func (s *stack) addToStackBottom(win fynedesk.Window) {
	s.clients = append(s.clients, win)
	s.mappingOrder = append(s.mappingOrder, win)
}

func (s *stack) getClients(clients []fynedesk.Window) []fynedesk.Window {
	return s.clients
}

func (s *stack) getMappingOrder() []fynedesk.Window {
	return s.mappingOrder
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
