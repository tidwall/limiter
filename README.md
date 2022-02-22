<p align="center">
<img 
    src="logo.png" 
    width="350" height="65" border="0" alt="limiter">
<br>
<a href="https://godoc.org/github.com/tidwall/limiter"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>
</p>

Limiter is a Golang library for limiting work coming from any number of goroutines.
This is useful when you need limit the maximum number of concurrent calls to a specific operation.

This library has two types:
- [Limiter](#limiter): Limits the number concurrent operations.
- [Queue](#queue): Queues limiter operations. Push/pop inputs/outputs.

## Install

Requires Go 1.18+

``` sh
go get github.com/tidwall/limiter
```

## Examples

### Limiter

``` go
package main

import (
	"io/ioutil"
	"net/http"

	"github.com/tidwall/limiter"
)

func main() {

	// Create a limiter for a maximum of 10 concurrent operations
	l := limiter.New(10)

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		input, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		defer r.Body.Close()

		var result []byte
		func() {
			l.Begin()
			defer l.End()
			// Do some intensive work here. It's guaranteed that only a
			// maximum of ten of these operations will run at the same time.
			result = []byte("rad!")
		}()

		w.Write(result.([]byte))
	})

	http.ListenAndServe(":8080", nil)
}
```

### Queue

```go
package main

import (
	"github.com/tidwall/limiter"
)

func main() {
	// Create a queue for a maximum of 10 concurrent operations. 
	q := limiter.NewQueue(10,
		func(in int) (out int) {
			// Do some intensive work here. It's guaranteed that only a maximum
			// of ten of these operations will run at the same time. 
			// Here we'll just multiple the input by 100 and return that value
			// as the output.
			return in * 100
		},
	)
	
	// Push 100 inputs onto the queue.
	for i := 0; i < 100; i++ {
		q.Push(i)
		// Once an input is pushed onto the queue, you can freely try to pop 
		// the output. The Pop method will return outputs in the same order
		// as their repsective inputs were pushed. It's possible that the 
		// next output is not ready because the background operation that is 
		// processing the input has yet to complete.
		for {
			out, ok := q.Pop()
			if !ok {
				// The next input has not completed or the queue is empty.
				break
			}
			println(out)
		}
	}

	// Finally you can use PopWait to wait on the next output. 
	for {
		out, ok := q.PopWait()
		if !ok {
			// The queue is empty
			break
		}
		println(out)
	}
	// output:
	// 0
	// 100
	// 200
	// ....
	// 9700
	// 9800
	// 9900
}
```

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

Limiter source code is available under the MIT [License](/LICENSE).




