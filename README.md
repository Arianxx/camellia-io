```
   _____                     _ _ _       
  / ____|                   | | (_)      
 | |     __ _ _ __ ___   ___| | |_  __ _ 
 | |    / _` | '_ ` _ \ / _ \ | | |/ _` |
 | |___| (_| | | | | | |  __/ | | | (_| |
  \_____\__,_|_| |_| |_|\___|_|_|_|\__,_|                                                                            
```
![](https://img.shields.io/github/license/Arianxx/camellia-io.svg)

Camellia is a efficient and easy-to-use single thread eventloop framework. It used to quickly create IO intensive application.
## Install
```bash
go get github.com/Arianxx/camellia-io
```

## Features

- Simple API.
- Low memory usage.
- Runs in a single thread but is non-blocking.
- Support for multiple task types.

## Examples
Simple echo server:
```golang
package main

import (
	"fmt"
	"github.com/arianxx/camellia-io"
)

func whenServing(el *camellia.EventLoop, _ *interface{}) {
	fmt.Println("Server start...")
}

func whenAccept(el *camellia.EventLoop, dataPtr *interface{}) {
	data := (*dataPtr).([]string)
	fmt.Println("Accept: ", data)
}

func echo(el *camellia.EventLoop, connPtr *interface{}) {
	conn := (*connPtr).(*camellia.Conn)
	msg := conn.Read()
	fmt.Println("Recv: ", string(msg))
	conn.Write(msg)
}

func main() {
	event := camellia.Event{
		Serving: whenServing,
		Open:    whenAccept,
		Data:    echo,
	}

	server := camellia.NewServer()
	lis, err := camellia.NewListener("tcp4", "127.0.0.1:12131", server.El)
	if err != nil {
		panic(err)
	}
	server.AddListener(lis)
	server.AddEvent(&event)

	if err := server.StartServe(); err != nil {
		panic(err)
	}
}
```

Period task:
```golang
package main

import (
	"fmt"
	"time"

	"github.com/arianxx/camellia-io"
)

var count = 0

func period(el *camellia.EventLoop, _ *interface{}) {
	fmt.Printf("%d seconds passed\n", count)
	count += 3
}

func main() {
	t := camellia.PeriodTask{
		Event:    period,
		Interval: 3 * time.Second,
	}

	server := camellia.NewServer()
	server.AddPeriodTask(&t)

	err := server.StartServe()
	if err != nil {
		panic(err)
	}
}
```

## Contact
[ArianX](https://github.com/arianxx)

## License
`Camellia-io` source code is available under the MIT License.