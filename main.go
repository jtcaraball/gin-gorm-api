package main

import "gin-gorm-api/server"

func main() {
	r := server.NewEngine()
	if err := r.Run(); err != nil {
		panic(err)
	}
}
