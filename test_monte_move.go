package main

import (
	"encoding/json"
	"log"
	"os"
)

func test_monte_move() {

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

	log.Println(move(state).Move)

}
