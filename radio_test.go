package pubsub

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

var gwg sync.WaitGroup
var gcount = 100
var ncount = 10000
var stop = []byte(strconv.Itoa(ncount - 1))

func waiter(r *Radio) {
	wg := r.Horn()
	for {
		if bytes.Equal(r.Event(), stop) {
			gwg.Done()
			return
		}
		wg.Wait()
		wg = r.Horn()
	}
}

func TestRadio_WaitSync(t *testing.T) {
	r := NewRadio(true)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r)
	}

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast([]byte(strconv.Itoa(i)))
	}
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
	gwg.Wait()
	r.Close()
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
	r := NewRadio(false)
	gwg.Add(gcount)
	for i := 0; i < gcount; i++ {
		go waiter(r)
	}

	t1 := time.Now()
	for i := 0; i < ncount; i++ {
		r.Broadcast([]byte(strconv.Itoa(i)))
	}
	fmt.Println("broadcast cost time :", time.Now().Sub(t1))
	gwg.Wait()
	r.Close()
}
