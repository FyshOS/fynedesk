package dbus

import (
	"container/list"
	"sync"
)

// NewSequentialSignalHandler returns an instance of a new
// signal handler that guarantees sequential processing of signals. It is a
// guarantee of this signal handler that signals will be written to
// channels in the order they are received on the DBus connection.
func NewSequentialSignalHandler() SignalHandler {
	return &sequentialSignalHandler{}
}

type sequentialSignalHandler struct {
	mu      sync.RWMutex
	closed  bool
	signals []*sequentialSignalChannelData
}

func (sh *sequentialSignalHandler) DeliverSignal(intf, name string, signal *Signal) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	if sh.closed {
		return
	}
	for _, scd := range sh.signals {
		scd.deliver(signal)
	}
}

func (sh *sequentialSignalHandler) Terminate() {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.closed {
		return
	}

	for _, scd := range sh.signals {
		scd.close()
		close(scd.ch)
	}
	sh.closed = true
	sh.signals = nil
}

func (sh *sequentialSignalHandler) AddSignal(ch chan<- *Signal) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.closed {
		return
	}
	sh.signals = append(sh.signals, &sequentialSignalChannelData{
		queue: list.New(),
		ch:    ch,
		done:  make(chan struct{}),
	})
}

func (sh *sequentialSignalHandler) RemoveSignal(ch chan<- *Signal) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if sh.closed {
		return
	}
	for i := len(sh.signals) - 1; i >= 0; i-- {
		if ch == sh.signals[i].ch {
			sh.signals[i].close()
			copy(sh.signals[i:], sh.signals[i+1:])
			sh.signals[len(sh.signals)-1] = nil
			sh.signals = sh.signals[:len(sh.signals)-1]
		}
	}
}

type sequentialSignalChannelData struct {
	stateLock sync.Mutex
	writeLock sync.Mutex
	queue     *list.List
	wg        sync.WaitGroup
	ch        chan<- *Signal
	done      chan struct{}
}

func (scd *sequentialSignalChannelData) deliver(signal *Signal) {
	// Avoid blocking the main DBus message processing routine;
	// queue signal to be dispatched later.
	scd.stateLock.Lock()
	scd.queue.PushBack(signal)
	scd.stateLock.Unlock()

	scd.wg.Add(1)
	go scd.deferredDeliver()
}

func (scd *sequentialSignalChannelData) deferredDeliver() {
	defer scd.wg.Done()

	// Ensure only one goroutine is in this section at once, to
	// make sure signals are sent over ch in the order they
	// are in the queue.
	scd.writeLock.Lock()
	defer scd.writeLock.Unlock()

	scd.stateLock.Lock()
	elem := scd.queue.Front()
	scd.queue.Remove(elem)
	scd.stateLock.Unlock()

	select {
	case scd.ch <- elem.Value.(*Signal):
	case <-scd.done:
	}
}

func (scd *sequentialSignalChannelData) close() {
	close(scd.done)
	scd.wg.Wait() // wait until all spawned goroutines return
}
