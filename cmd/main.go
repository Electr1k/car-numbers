package main

import "car-numers/config"

func main() {
	cfg := config.MustLoad()

	println(cfg)
}
