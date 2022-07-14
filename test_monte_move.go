package main

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

func test_monte_move(t *testing.T) {

	test_body, err := os.ReadFile("test_request.json")
	if err != nil {
		t.Fail()
		panic(err.Error())
	}

	state := GameState{}

	err2 := json.Unmarshal(test_body, &state)

	if err2 != nil {
		t.Fail()
		log.Printf("ERROR: Failed to decode move json, %s", err)
		return
	}

	log.Println(move(state).Move)

	t.Fatal()

}
