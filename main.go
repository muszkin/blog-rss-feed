package main

import (
	"fmt"
	config2 "github.com/muszkin/blog-rss-feed/internal/config"
)

func main() {
	config, err := config2.Read()
	if err != nil {
		fmt.Printf("Cannot read config file: %v", err)
	}
	config.SetUser("muszkin")
	config, err2 := config2.Read()
	if err2 != nil {
		fmt.Printf("Cannot read config file: %v", err2)
	}
	fmt.Println(config)
}
