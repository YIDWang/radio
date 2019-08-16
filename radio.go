package pubsub

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Radio publishes the latest event
// Every listener will be notified,it can get the latest event through the Event
type Radio struct {
	lock        sync.RWMutex
	latestEvent []byte
	horn        unsafe.Pointer //*discard.WaitGroup

	echan   chan []byte
	discard bool
}

// NewRadio create a radio
func NewRadio(discard bool) *Radio {
	if discard {
		r := &Radio{
			horn:    horn(),
			echan:   make(chan []byte, 1),
			discard: discard,
		}
		go r.broadcast()
		return r
	}
	return &Radio{
		horn:    horn(),
		discard: discard,
	}
}

// Horn return a horn
func (r *Radio) Horn() *sync.WaitGroup {
	return (*sync.WaitGroup)(atomic.LoadPointer(&r.horn))
}

// Event return the latest event
func (r *Radio) Event() []byte {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.latestEvent
}

// Broadcast update local events and publish notifications
func (r *Radio) Broadcast(event []byte) {
	if !r.discard {
		r.publish(event)
		return
	}
	for {
		select {
		case r.echan <- event:
			return
		case ev := <-r.echan:
			if ev == nil {
				return
			}
		}
	}
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

func (r *Radio) publish(event []byte) {
	r.lock.Lock()
	r.latestEvent = event
	r.Horn().Done()
	atomic.StorePointer(&r.horn, horn())
	r.lock.Unlock()
}

// Close shut dwon radio
func (r *Radio) Close() {
	if r.discard {
		close(r.echan)
	}
}

// horn get a horn
func horn() unsafe.Pointer {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return (unsafe.Pointer)(wg)
}
