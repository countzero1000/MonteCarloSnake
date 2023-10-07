package main

import (
	"log"
)

func info() BattlesnakeInfoResponse {
	log.Println("INFO")
	return BattlesnakeInfoResponse{
		APIVersion: "1",
		Author:     "",
		Color:      "#12d5db",
		Head:       "shades",
		Tail:       "sharp",
	}
}

func start(state GameState) {
	log.Printf("%s START\n", state.Game.ID)
}

func end(state GameState) {
	log.Printf("%s END\n\n", state.Game.ID)
}

func move(state GameState) BattlesnakeMoveResponse {

	tree := new_tree(state)

	return BattlesnakeMoveResponse{
		Move: tree.monte_move().Move,
	}
}
