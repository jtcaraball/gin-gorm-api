package main

import (
	"gin-gorm-api/handler"
	"gin-gorm-api/server"
)

func main() {
	config, err := server.LoadConfig()
	if err != nil {
		panic(err)
	}

	r, err := server.NewEngine(config, handler.Greeter{})
	if err != nil {
		panic(err)
	}

	if err = r.Run(); err != nil {
		panic(err)
	}
}
