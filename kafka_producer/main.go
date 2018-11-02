package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"strconv"
	"time"
)

var globalcounter = 0

func main() {
	for {
		globalcounter += 1
		time.Sleep(1*time.Second)
		sentMessage()
	}
}

func sentMessage() {
	topic := "microbases"

	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		panic(err)
	}
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	message := "The counter is now at: " + strconv.Itoa(globalcounter)
	conn.WriteMessages(
		kafka.Message{Value: []byte(message)},
	)

	println(message)
	conn.Close()
}
