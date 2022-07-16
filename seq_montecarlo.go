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
	name   string
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

const c float64 = 2

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
		name:   game.You.Name,
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
	tree.root.expandNode()
	println("running with", iterations, "iterations")
	for i := 0; i < iterations; i++ {
		tree.expand_tree()
	}

	return tree.root.select_best_move(tree.player, tree.name)
}

func (node *Node) select_best_move(snake_id string, name string) rules.SnakeMove {

	best_move := rules.MoveRight

	var most_val float32 = 0

	for _, child := range node.children {

		println(child.action.Move, child.sims, child.wins)

		val := (float32)(child.wins)
		if val > most_val {
			most_val = val
			best_move = child.action.Move
		}
	}

	println(name, "selected best move", best_move)
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
		return math.MaxInt
	}
	discover := (c * math.Sqrt(math.Log((float64)(parent_sims+1))/(float64)(sims)))
	reward := ((float64)(wins) / (float64)(sims))

	// println("reward", reward, "discover", discover, wins, sims, parent_sims)
	return reward + discover

}

func (node *Node) play_out() {
	// println("playing out")
	iterations := 0
	game_over := false
	// snake := node.board.board.Snakes[0]
	// println(snake.Health, "starting health", snake.EliminatedCause, snake.Body[0].X, snake.Body[0].Y)
	copy_board := node.board.copy()
	for !game_over {

		if iterations+copy_board.board.Turn >= 250 {
			game_over = true
			break
		}

		selected_move := []rules.SnakeMove{}

		for _, snake := range node.board.board.Snakes {
			moves := node.board.getValidMoves(snake.ID)
			// for _, move := range moves {
			// 	// println("valid move", move.Move)
			// }
			if len(moves) == 0 {
				moves = []rules.SnakeMove{{ID: snake.ID, Move: rules.MoveUp}}
			}
			move := moves[rand.Intn(len(moves))]
			// println("applied move", move.Move)
			selected_move = append(selected_move, move)
			// println("helf", snake.Health)
		}
		new_game_over, _, err := copy_board.executeActions(selected_move)
		// copy_board.board = *new_board
		game_over = new_game_over

		if err != nil {
			println(err.Error())
			panic("error thrown while playing out")
		}
		iterations += 1
	}
	// println("finished playout with iterations", iterations)
	// snake = copy_board.board.Snakes[0]
	// println(snake.Health, "ending health", snake.EliminatedCause, snake.Body[0].X, snake.Body[0].Y)

	winner := get_winner(copy_board.board.Snakes)
	node.back_prop(winner)

}

func get_winner(snakes []rules.Snake) string {

	for _, snake := range snakes {
		if len(snake.EliminatedCause) == 0 {
			// println(snake.Health, "winner helf")
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
