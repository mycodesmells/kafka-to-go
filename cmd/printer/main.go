package main

import "fmt"
import "github.com/mycodesmells/kafka-to-go/kafka"

func main() {
	consumer, err := kafka.NewConsumer("group", "example", "localhost:2181")
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Kafka consumer: %v", err))
	}

	err = consumer.Start()
	if err != nil {
		panic(fmt.Sprintf("Failed to start Kafka consumer: %v", err))
	}
}
