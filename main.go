package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Timestamp string `json:"timestamp"`
}

func main() {
	// Connect to Redis server
	rdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.221.142:6379",
		Password: "sOmE_sEcUrE_pAsS",
		DB:       0,
	})

	connStr := "user=postgres dbname=backend_my_house sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// There is no error because go-redis automatically reconnects on error.
	pubsub := rdb.Subscribe("myCoolChannel1")

	// Close the subscription when we are done.
	defer pubsub.Close()
	ch := pubsub.Channel()

	var loop chan struct{}

	go func() {
		for msg := range ch {
			var data map[string]interface{}
			err := json.Unmarshal([]byte(msg.Payload), &data)
			if err != nil {
				fmt.Println(err)
			}

			selectString := fmt.Sprintf("SELECT id, email, phone FROM users WHERE id = %v", data["user_id"])
			rows, err := db.Query(selectString)
			if err != nil {
				panic(err)
			}

			users := []User{}
			for rows.Next() {
				user := User{}
				err := rows.Scan(&user.ID, &user.Email, &user.Phone)
				if err != nil {
					panic(err)
				}
				user.Timestamp = time.Now().Format(time.RFC1123)
				users = append(users, user)
			}

			b, err := json.Marshal(users)
			if err != nil {
				fmt.Println(err)
			}

			rdb.Publish("myCoolChannel2", b)
			rows.Close()
		}
	}()

	log.Printf(" [*] Waiting for requests. To exit press CTRL+C")
	<-loop
}
