package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"weather-server/src/repository"
	"weather-server/src/routes"
)

func main() {
	fmt.Println("ServerStart")
	err := repository.InitRedis("localhost:"+os.Getenv("REDIS_PORT"), os.Getenv("REDIS_PASSWORD"))
	if err != nil {
		log.Fatalf("Critical error: couldn't connect to Redis: %v", err)
	}

	router := routes.NewRouter()

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
