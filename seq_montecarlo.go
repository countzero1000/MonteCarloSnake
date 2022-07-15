package main

import (
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/joho/godotenv"
)

type Tree struct {
	player string
	root   *Node
}

type Node struct {
	player       string
	children     []*Node
	parent       *Node
	board        Simulation
	sims         int
	wins         int
	action       rules.SnakeMove
	player_order map[string]int
	player_arr   []string
}

const c float64 = 1.41

func new_tree(game GameState) Tree {
	println("making new tree for", game.You.ID)
	player_order := make(map[string]int)
	player_arr := []string{}
	player_arr = append(player_arr, game.You.ID)
	for _, snake := range game.Board.Snakes {
		if snake.ID != game.You.ID {
			player_arr = append(player_arr, snake.ID)
		}
	}

	println("array for", game.You.ID, "=================")
	for _, id := range player_arr {
		println(id)
	}
	println("=========================")

	for i, snake := range player_arr {
		player_order[snake] = i
	}
	tree := Tree{
		player: game.You.ID,
		root: &Node{
			player_arr:   player_arr,
			player_order: player_order,
			children:     []*Node{},
			parent:       nil,
			board:        simulationFromGame(&game),
			wins:         0,
			sims:         0,
		},
	}
	tree.root.player = tree.root.get_prev_player(game.You.ID)
	println("previous player for", game.You.ID, "is", tree.root.player)
	return tree
}

func (node *Node) get_next_player(snake_id string) string {
	order := node.player_order[snake_id]
	return node.player_arr[(order+1)%len(node.player_arr)]
}
func (node *Node) get_prev_player(snake_id string) string {
	order := node.player_order[snake_id]
	if order-1 < 0 {
		return node.player_arr[len(node.player_order)-1]
	}
	return node.player_arr[(order-1)%len(node.player_arr)]
}

func (tree *Tree) monte_move() rules.SnakeMove {

	tree.root.expandNode()
	godotenv.Load(".env")
	iterations, err := strconv.Atoi(os.Getenv("iterations"))
	if err != nil {
		println(err.Error())
		panic("error")
	}
	println("running with", iterations, "iterations")
	for i := 0; i < iterations; i++ {
		tree.expand_tree()
	}

	return tree.root.select_best_move(tree.player)
}

func (node *Node) select_best_move(snake_id string) rules.SnakeMove {

	best_move := rules.MoveRight
	best_node := node.children[0]

	var most_val float32 = 0

	for _, child := range node.children {

		println(child.action.Move, child.sims)

		val := (float32)(child.sims)
		if val > most_val {
			most_val = val
			best_node = child
			best_move = child.action.Move
		}
	}

	println(node.player, "selected best move with", most_val, "action", best_move, "with location", best_node.board.board.Snakes[0].Body[0].X, best_node.board.board.Snakes[0].Body[0].Y)
	return rules.SnakeMove{ID: snake_id, Move: best_move}
}

func (tree *Tree) expand_tree() {
	var promising_node = tree.root.select_node()

	promising_node.expandNode()

	var test_node = promising_node

	if len(promising_node.children) > 0 {
		test_node = promising_node.children[rand.Intn(len(promising_node.children))]
	}

	test_node.play_out()
}

func (node *Node) expandNode() {

	new_player := node.get_next_player(node.player)
	move_matrix := node.board.getValidMoves(new_player)

	// println("expanding for", new_player, "after", node.player)

	for _, joint_move := range move_matrix {
		node.children = append(node.children, create_child(node, joint_move, node.board, new_player))
	}

}

func (node *Node) select_node() *Node {

	if len(node.children) > 0 {
		var max_val float64 = 0
		best_node := node.children[0]
		for _, child := range node.children {
			parent_sims := 1
			if node.parent != nil {
				parent_sims = node.parent.sims
			}
			val := calc_utc_val(child.wins, child.sims, parent_sims)
			if val > max_val {
				max_val = val
				best_node = child
			}
		}
		return best_node.select_node()
	}
	return node
}

func calc_utc_val(wins int, sims int, parent_sims int) float64 {
	if sims == 0 {
		return math.Inf(1)
	}
	return ((float64)(wins) / (float64)(sims)) + (c * math.Pow(math.Log((float64)(parent_sims))/(float64)(sims), .5))

}

func add_to_map(m map[string]int, key string, insert_val int) {
	_, exists := m[key]
	if exists {
		m[key] += insert_val
	} else {
		m[key] = insert_val
	}

}

func get_move_by_snake(snake_id string, joint_move []rules.SnakeMove) rules.SnakeMove {
	for _, move := range joint_move {
		if move.ID == snake_id {
			return move
		}
	}
	return rules.SnakeMove{ID: snake_id, Move: rules.MoveDown}
}

func (node *Node) play_out() {
	// println("playing out")
	iterations := 0
	game_over := false
	copy_board := node.board.copy()
	for !game_over {

		joint_moves := copy_board.generateMoveMatrix()
		if len(joint_moves) == 0 {
			game_over = true
			break
		}
		selected_move := joint_moves[rand.Intn(len(joint_moves))]
		new_game_over, new_board, err := copy_board.executeActions(selected_move)
		copy_board.board = *new_board
		game_over = new_game_over

		if err != nil {
			println(err.Error())
			panic("error thrown while playing out")
		}
		iterations += 1
	}

	winner := get_winner(copy_board.board.Snakes)
	node.back_prop(winner)

}

func get_winner(snakes []rules.Snake) string {

	for _, snake := range snakes {
		if len(snake.EliminatedCause) == 0 {
			return snake.ID
		}

	}
	return ""
}

func (node *Node) back_prop(winner string) {

	if node.player == winner {
		node.wins += 1
	}
	node.sims += 1

	if node.parent != nil {
		node.parent.back_prop(winner)
	}
}

func create_child(parent *Node, action rules.SnakeMove, board Simulation, player string) *Node {

	board_copy := board.copy()
	_, new_board, _ := board_copy.executeAction(action)
	board_copy.board = *new_board

	return &Node{
		player_order: parent.player_order,
		player_arr:   parent.player_arr,
		sims:         0,
		wins:         0,
		action:       action,
		children:     []*Node{},
		parent:       parent,
		board:        board_copy,
		player:       player,
	}
}
