package main

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

func Test_MonteCarlo(t *testing.T) {

	test_body, err := os.ReadFile("test_request.json")
	if err != nil {
		panic(err.Error())
	}

	state := GameState{}

	err2 := json.Unmarshal(test_body, &state)

	if err2 != nil {
		log.Printf("ERROR: Failed to decode move json, %s", err)
		return
	}

	move(state)
}
