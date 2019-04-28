package main

import (
	"fmt"
	"github.com/arianxx/camellia-io"
	"time"
)

func serving(el *camellia.EventLoop, _ *interface{}) {
	fmt.Println("Server start...")
}

func accept(el *camellia.EventLoop, dataPtr *interface{}) {
	data := (*dataPtr).([]string)
	fmt.Println("Accept: ", data)
}

func echo(el *camellia.EventLoop, connPtr *interface{}) {
	conn := (*connPtr).(*camellia.Conn)
	msg := conn.Read()
	fmt.Println("Recv: ", string(msg))
	conn.Write(msg)
}

func peridEcho(el *camellia.EventLoop, _ *interface{}) {
	fmt.Println("1s elapsed")
}

func main() {
	event := camellia.Event{
		Serving: serving,
		Open:    accept,
		Data:    echo,
	}
	loop := camellia.NewEventLoop()
	loop.AddEvent(&event)

	task := &camellia.PeriodTask{
		Interval: time.Second,
		Event:    peridEcho,
	}
	loop.AddPeriodTask(task)

	lis, err := camellia.NewListener("tcp4", "127.0.0.1:12131", loop)
	if err != nil {
		panic(err)
	}

	err = lis.RegisterAccept()
	if err != nil {
		panic(err)
	}

	loop.Run()
}
