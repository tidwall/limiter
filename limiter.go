package limiter

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
