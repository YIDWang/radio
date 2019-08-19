package pubsub

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	invalidEvent = unsafe.Pointer(new(interface{}))
	initialID    int64
)

func idGenerator() int64 {
	return atomic.AddInt64(&initialID, 1)
}

// Event broadcast on the radio
// Id guarantees the order of events
type Event struct {
	content interface{}
	id      int64
}

// NewEvent return ordered event
func NewEvent(content interface{}) *Event {
	return &Event{
		content: content,
		id:      idGenerator(),
	}
}

// Content event content
func (ev *Event) Content() interface{} {
	return ev.content
}

// Radio publishes the latest event
// Every listener will be notified,it can get the latest event through the Event
type Radio struct {
	lock        sync.RWMutex
	latestEvent unsafe.Pointer
	horn        unsafe.Pointer //*discard.WaitGroup

	echan   chan unsafe.Pointer
	discard bool
	wg      sync.WaitGroup
}

// NewRadio create a radio
func NewRadio(discard bool) *Radio {
	r := &Radio{
		horn:        horn(),
		discard:     discard,
		latestEvent: invalidEvent,
	}
	if discard {
		r.echan = make(chan unsafe.Pointer, 1)
		r.wg.Add(1)
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
		nr.wg.Add(1)
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
		close(r.echan)
		r.wg.Wait()
	} else {
		r.echan = make(chan unsafe.Pointer, 1)
		r.wg.Add(1)
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
func (r *Radio) Event() *Event {
	ev := atomic.LoadPointer(&r.latestEvent)
	if ev == invalidEvent {
		return nil
	}
	return (*Event)(ev)
}

// Broadcast update local events and publish notifications
func (r *Radio) Broadcast(event *Event) {
	if event == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if !r.discard {
		if event.id < r.Event().id {
			return
		}
		r.publish(unsafe.Pointer(event))
		return
	}
	for {
		select {
		case r.echan <- unsafe.Pointer(event):
			return
		case ev := <-r.echan:
			if ev == nil {
				return
			}
			if event.id < (*Event)(ev).id {
				event = (*Event)(ev)
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
	defer r.wg.Done()
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
