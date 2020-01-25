package notify

// ScreenChangeNotify is an interface that can be used by objects interested in screen configuration changes
type ScreenChangeNotify interface {
	ScreenChangeNotify()
}
