package limiter

import "sync"

// Limiter is for limiting the number of concurrent operations. This
type Limiter struct {
	limit int
	cond  *sync.Cond
	count int
}

// New returns a new Limiter. The limit param is the maximum number of
// concurrent operations.
func New(limit int) *Limiter {
	return &Limiter{
		limit: limit,
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}

// Begin an operation.
func (l *Limiter) Begin() {
	l.cond.L.Lock()
	for l.count == l.limit {
		l.cond.Wait()
	}
	l.count++
	l.cond.L.Unlock()
}

// End the operation.
func (l *Limiter) End() {
	l.cond.L.Lock()
	l.count--
	l.cond.Broadcast()
	l.cond.L.Unlock()
}
