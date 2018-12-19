<p align="center">
<img 
    src="logo.png" 
    width="350" height="65" border="0" alt="GJSON">
<br>
<a href="https://godoc.org/github.com/tidwall/limiter"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>
</p>

Limiter is a Golang library for limiting work coming from any number of goroutines.
This is useful when you need limit the maximum number of concurrent calls to a specific operation.

## Install

``` sh
go get github.com/tidwall/limiter
```


## Example

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
			// Do some intensive work here. It's guaranteed that only a maximum of ten
			// of these operations will run at the same time.
			result = []byte("rad!")
		}()

		w.Write(result.([]byte))
	})

	http.ListenAndServe(":8080", nil)
}
```

## Contact

Josh Baker [@tidwall](http://twitter.com/tidwall)

## License

Limiter source code is available under the MIT [License](/LICENSE).




