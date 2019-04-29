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
