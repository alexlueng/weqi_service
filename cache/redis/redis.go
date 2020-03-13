package redis

import (
	"fmt"
	"github.com/go-redis/redis"
)

var Client *redis.Client

func RedisClient(connString string) {
	client := redis.NewClient(&redis.Options{
		Addr: connString,
		Password: "",
		DB: 0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error while connecting redis: ", err)
		return
	}
	fmt.Println(pong)
	Client = client
}