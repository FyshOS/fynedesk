package notify

// DesktopNotify allows modules to be informed when user changes virtual desktop
type DesktopNotify interface {
	DesktopChangeNotify(int)
}
