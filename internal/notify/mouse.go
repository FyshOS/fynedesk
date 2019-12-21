package notify

type MouseNotify interface {
	MouseInNotify(x int, y int)
	MouseOutNotify()
}
