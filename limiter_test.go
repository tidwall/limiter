package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	l := New(5)
	var wg sync.WaitGroup
	var count int32
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Begin()
			defer l.End()
			n := atomic.AddInt32(&count, 1)
			defer atomic.AddInt32(&count, -1)
			if n < 0 || n > 5 {
				t.Fatal("out of range")
			}
			time.Sleep(time.Millisecond)
		}()
	}
	wg.Wait()
}
