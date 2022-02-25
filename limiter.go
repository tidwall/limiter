package limiter

import "sync"

// Limiter is for limiting the number of concurrent operations.
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

type queueItem[I, O any] struct {
	in   I
	out  O
	ok   bool
	prev *queueItem[I, O]
	next *queueItem[I, O]
}

// Queue is a limiter queue operations that executes each operation in
// background goroutines, where each operation has a single input and output.
// The inputs are pushed onto the queue using Push, and the output can be
// retrieved using Pop.
type Queue[I, O any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	l    *Limiter
	head *queueItem[I, O]
	tail *queueItem[I, O]
	op   func(I) O
}

// NewQueue returns a limiter queue.
func NewQueue[I, O any](limit int, op func(in I) (out O)) *Queue[I, O] {
	q := &Queue[I, O]{l: New(limit)}
	q.cond = sync.NewCond(&q.mu)
	q.op = op
	q.head = new(queueItem[I, O])
	q.tail = new(queueItem[I, O])
	q.head.next = q.tail
	q.tail.prev = q.head
	return q
}

// Push an input onto the queue for background processing.
func (q *Queue[I, O]) Push(in I) {
	item := new(queueItem[I, O])
	item.in = in
	// push to tail
	q.mu.Lock()
	q.tail.prev.next = item
	item.prev = q.tail.prev
	item.next = q.tail
	q.tail.prev = item
	q.mu.Unlock()
	// execute operation
	go func() {
		q.l.Begin()
		defer func() {
			q.l.End()
			q.mu.Lock()
			item.ok = true
			q.cond.Broadcast()
			q.mu.Unlock()
		}()
		item.out = q.op(in)
	}()
}

func (q *Queue[I, O]) pop() (out O, ok bool) {
	if q.head.next.ok {
		item := q.head.next
		out, ok = item.out, true
		q.head.next = item.next
		item.next.prev = item.prev
	}
	return out, ok
}

// Pop output off the queue. The outputs will be returned in order of their
// respective inputs. Returns false if the queue is empty of if the next input
// operation has not yet finished processing.
func (q *Queue[I, O]) Pop() (out O, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pop()
}

// PopWait works like Pop but it wait for the next input operation to finish
// processing before returning its respective output. Returns false if the
// queue is empty.
func (q *Queue[I, O]) PopWait() (out O, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		out, ok = q.pop()
		if ok {
			// we have an item
			return out, ok
		}
		if q.head.next == q.tail {
			// the queue is empty
			return out, false
		}
		q.cond.Wait()
	}
}

type result[O any] struct {
	err error
	out O
}

// Group for running background operations.
type Group[I, O any] struct {
	wg sync.WaitGroup
	q  *Queue[I, result[O]]
}

// NewGroup returns a Group for running background operations.
func NewGroup[I, O any](limit int, op func(in I) (out O, err error)) *Group[I, O] {
	g := new(Group[I, O])
	g.q = NewQueue[I, result[O]](limit, func(in I) result[O] {
		var res result[O]
		if op != nil {
			res.out, res.err = op(in)
		}
		return res
	})
	return g
}

// Drain all pending outputs.
// This acts as a barrier to ensure that there are no more group operations
// running in the background.
func (g *Group[I, O]) Drain() {
	for {
		_, ok := g.q.PopWait()
		if !ok {
			return
		}
		g.wg.Done()
	}
}

// Send an input to the group for background processing.
func (g *Group[I, O]) Send(in I) {
	g.wg.Add(1)
	g.q.Push(in)
}

// Recv receives pending outputs. Setting "wait" to true will make this
// function wait for all inputs to complete being processed before returning.
// The "results" callback will fire for all outputs in the same order as their
// respective inputs.
// If the group operation or callback returned an error then the iterator
// will stop and that error will be returned to the call of this function.
func (g *Group[I, O]) Recv(wait bool, results func(out O) error) error {
	for {
		var tup result[O]
		var ok bool
		if wait {
			tup, ok = g.q.PopWait()
		} else {
			tup, ok = g.q.Pop()
		}
		if !ok {
			return nil
		}
		g.wg.Done()
		if tup.err != nil {
			return tup.err
		}
		if results != nil {
			if err := results(tup.out); err != nil {
				return err
			}
		}
	}
}
