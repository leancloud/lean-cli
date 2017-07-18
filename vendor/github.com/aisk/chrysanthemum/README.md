# chrysanthemum

[![GoDoc](https://godoc.org/github.com/aisk/chrysanthemum?status.svg)](https://godoc.org/github.com/aisk/chrysanthemum)

菊花【Jú Huā】

![](http://www.chrysanthemums.org/wp-content/uploads/2016/05/artificial-white-chrysanthemum-flowers.png)

Add a text spinner to your terminal applications.

## Usage

```go
package main

import (
	"github.com/aisk/chrysanthemum"
	"time"
)

func main() {
	c := chrysanthemum.New("I'll be ok").Start()
	time.Sleep(5 * time.Second)
	c.End()
	c = chrysanthemum.New("I'll be error").Start()
	time.Sleep(5 * time.Second)
	c.Failed()
}
```
