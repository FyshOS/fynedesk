package notify

// MouseNotify is an interface that can be used by objects interested in when the mouse enters or exits the desktop
type ScreenChangeNotify interface {
	ScreenChangeNotify()
}
