package limiter

import (
	"sort"
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
				panic("out of range")
			}
			time.Sleep(time.Millisecond)
		}()
	}
	wg.Wait()
}

func TestQueue(t *testing.T) {
	N := 100_000
	q := NewQueue(5,
		func(in int) (out int) {
			return in * 10
		},
	)
	var outs []int
	for i := 0; i < N; i++ {
		q.Push(i)
		out, ok := q.Pop()
		if ok {
			outs = append(outs, out)
		}
	}
	for {
		out, ok := q.PopWait()
		if !ok {
			break
		}
		outs = append(outs, out)
	}
	if len(outs) != N {
		t.Fatalf("expected %d got %d", N, len(outs))
	}
	if !sort.IntsAreSorted(outs) {
		t.Fatal("out of order")
	}
	for i := 0; i < N; i++ {
		if outs[i] != i*10 {
			t.Fatalf("expected %d got %d", i*10, outs[i])
		}
	}
}
