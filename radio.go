package radio

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var InvalidEvent = unsafe.Pointer(new(interface{}))

// Radio publishes the latest event
// Every listener will be notified,it can get the latest event through the Event
type Radio struct {
	lock        sync.RWMutex
	latestEvent unsafe.Pointer
	horn        unsafe.Pointer //*discard.WaitGroup

	echan   chan unsafe.Pointer
	discard bool
}

// NewRadio create a radio
func NewRadio(discard bool) *Radio {
	r := &Radio{
		horn:        horn(),
		discard:     discard,
		latestEvent: InvalidEvent,
	}
	if discard {
		r.echan = make(chan unsafe.Pointer, 1)
		go r.broadcast()
	}
	return r
}

// CloneRadio clone radio by assigning the discard method
func CloneRadio(r *Radio, discard bool) *Radio {
	nr := &Radio{
		discard:     discard,
		horn:        atomic.LoadPointer(&r.horn),
		latestEvent: atomic.LoadPointer(&r.latestEvent),
	}
	if discard {
		nr.echan = make(chan unsafe.Pointer, 1)
		go nr.broadcast()
	}
	return nr
}

// Discard return discard
func (r *Radio) Discard() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.discard
}

// Transform change the current broadcast mode
func (r *Radio) Transform() {
	r.lock.Lock()
	if r.discard {
		select {
		case ev := <-r.echan:
			r.publish(ev)
		default:
		}
		close(r.echan)
	} else {
		r.echan = make(chan unsafe.Pointer, 1)
		go r.broadcast()
	}
	r.discard = !r.discard
	r.lock.Unlock()
}

// Listene return a audio listener
func (r *Radio) Listener() *sync.WaitGroup {
	return (*sync.WaitGroup)(atomic.LoadPointer(&r.horn))
}

// Event return the latest event
func (r *Radio) Event() interface{} {
	ev := atomic.LoadPointer(&r.latestEvent)
	if ev == InvalidEvent {
		return nil
	}
	return *(*interface{})(ev)
}

// Broadcast update local events and publish notifications
func (r *Radio) Broadcast(event interface{}) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if !r.discard {
		r.publish(unsafe.Pointer(&event))
		return
	}
	for {
		select {
		case r.echan <- unsafe.Pointer(&event):
			return
		case ev := <-r.echan:
			if ev == nil {
				return
			}
		}
	}
}

// Close shut dwon radio
func (r *Radio) Close() {
	r.lock.Lock()
	if r.discard {
		close(r.echan)
		r.discard = false
	}
	r.lock.Unlock()
}

func (r *Radio) broadcast() {
	for {
		select {
		case event := <-r.echan:
			if event == nil {
				return
			}
			r.publish(event)
		}
	}
}

func (r *Radio) publish(event unsafe.Pointer) {
	atomic.StorePointer(&r.latestEvent, event)
	oldHorn := atomic.SwapPointer(&r.horn, horn())
	(*sync.WaitGroup)(oldHorn).Done()
}

// horn get a horn
func horn() unsafe.Pointer {
	wg := sync.WaitGroup{}
	wg.Add(1)
	return (unsafe.Pointer)(&wg)
}
