package radio

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
var ncount int = 1000

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func waiter(r *Radio) {
	wg := r.Listener()
	for {
		ev := r.Event()
		switch ev.(type) {
		case nil:
		case int:
			if ev.(int) == (ncount - 1) {
				gwg.Done()
				return
			}
		}
		wg.Wait()
		wg = r.Listener()
	}
}

func TestRadio_WaitDiscard(t *testing.T) {
	r := NewRadio(false)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r)
	}

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast(i)
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
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r)
	}

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast(i)
	}
	gwg.Wait()
	r.Close()
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
}

func TestRadioTransform(t *testing.T) {
	r := NewRadio(false)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r)
	}
	go func() {
		for i := 0; i < 1000; i++ {
			r.Transform()
		}
	}()

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast(i)
	}
	gwg.Wait()
	r.Close()
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
}
