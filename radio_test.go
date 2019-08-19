package pubsub

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
)

var gwg sync.WaitGroup
var gcount = 500
var ncount int = 100

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func Println(r *Radio) {
	time.Sleep(time.Minute)
	fmt.Println(r.Event().Content().(int))
}

func waiter(r *Radio, index int) {
	wg := r.Listener()
	var ev *Event
	for {
		wg.Wait()
		wg = r.Listener()
		ev = r.Event()
		if ev.Content().(int) == (ncount - 1) {
			gwg.Done()
			return
		}
	}
}

func TestRadio_WaitDiscard(t *testing.T) {
	r := NewRadio(false)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r, i)
	}
	time.Sleep(time.Second)

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast(NewEvent(i))
	}
	gwg.Wait()
	r.Close()
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
}

func TestScan(t *testing.T) {
	var lock sync.Mutex
	t1 := time.Now()
	for i := 0; i < gcount; i++ {
		for j := 0; j < ncount; j++ {
			lock.Lock()
			lock.Unlock()
		}
	}
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
}

func TestRadio_Wait(t *testing.T) {
	r := NewRadio(true)
	go Println(r)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r, i)
	}
	time.Sleep(time.Second)

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast(NewEvent(i))
	}
	gwg.Wait()
	r.Close()
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
}

func TestRadioTransform(t *testing.T) {
	r := NewRadio(false)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r, i)
	}
	go func() {
		for i := 0; i < 1000; i++ {
			r.Transform()
		}
	}()
	time.Sleep(time.Second)

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast(NewEvent(i))
	}
	gwg.Wait()
	r.Close()
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
}
