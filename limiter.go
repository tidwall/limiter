package limiter

import "sync"

// Limiter is for limiting the number of concurrent operations. This
type Limiter struct{ sem chan struct{} }

// New returns a new Limiter. The limit param is the maximum number of
// concurrent operations.
func New(limit int) *Limiter {
	return &Limiter{make(chan struct{}, limit)}
}

// Begin an operation.
func (l *Limiter) Begin() {
	l.sem <- struct{}{}
}

// End the operation.
func (l *Limiter) End() {
	<-l.sem
}

// Group is for grouping operations together and allows for waiting for
// all operations to complete.
type Group struct {
	mu  sync.Mutex
	wg  sync.WaitGroup
	err error
	l   *Limiter
}

// NewGroup returns a limiter operation group.
func NewGroup(limit int) *Group {
	return &Group{l: New(limit)}
}

func (g *Group) Do(op func() error) error {
	g.wg.Add(1)
	go func() {
		g.l.Begin()
		defer func() {
			g.l.End()
			g.wg.Done()
		}()
		g.mu.Lock()
		if g.err != nil {
			g.mu.Unlock()
			return
		}
		g.mu.Unlock()
		err := op()
		if err != nil {
			g.mu.Lock()
			g.err = err
			g.mu.Unlock()
		}
	}()
	return nil
}

func (g *Group) Wait() error {
	g.wg.Wait()
	g.mu.Lock()
	err := g.err
	g.mu.Unlock()
	return err
}
