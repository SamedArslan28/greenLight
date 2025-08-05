package main

import (
	"github.com/joho/godotenv"
	"log"
	"sync"
)

var once sync.Once

// LoadEnvOnce loads .env file only once during tests.
func LoadEnvOnce(path string) {
	once.Do(func() {
		err := godotenv.Load(path)
		if err != nil {
			log.Fatalf("Failed to load .env file at %s: %v", path, err)
		}
	})
}
