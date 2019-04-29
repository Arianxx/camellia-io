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
