package main

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

func main() {
	// Connect to Redis server
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.221.142:6379",
		Password: "sOmE_sEcUrE_pAsS",
		DB:       0,
	})

	// There is no error because go-redis automatically reconnects on error.
	pubsub := rdb.Subscribe("myCoolChannel1")

	// Close the subscription when we are done.
	defer pubsub.Close()

	ch := pubsub.Channel()

	var loop chan struct{}

	go func() {
		for msg := range ch {
			fmt.Println(msg.Channel, msg.Payload)
			rdb.Publish("myCoolChannel2", msg.Payload)
		}
	}()

	log.Printf(" [*] Waiting for requests. To exit press CTRL+C")
	<-loop
}
