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

const c float64 = 1.141
const DEBUG_MODE = false

func new_tree(game GameState) Tree {
	player_order := make(map[string]int)
	player_arr := []string{}
	player_arr = append(player_arr, game.You.ID)
	for _, snake := range game.Board.Snakes {
		if snake.ID != game.You.ID {
			player_arr = append(player_arr, snake.ID)
		}
	}

	for _, id := range player_arr {
		println(id)
	}

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

	godotenv.Load(".env")
	iterations, err := strconv.Atoi(os.Getenv("iterations"))
	if err != nil {
		println(err.Error())
		panic("error")
	}
	tree.root.expandNode()
	for i := 0; i < iterations; i++ {
		tree.expand_tree()
	}

	return tree.root.select_best_move(tree.player, tree.name)
}

func (node *Node) recur_print() {

	if !DEBUG_MODE {
		return
	}

	snake := get_snake(node.board.board, node.action.ID).Body[0]
	println("player:", node.player, "moved", node.action.Move, "wins:", node.wins, "sims:", node.sims, "position", snake.X, snake.Y)
	printMap(&node.board.board)

	if len(node.children) == 0 {
		for _, snake := range node.board.board.Snakes {

			println("snake", snake.ID, "eliminated by", snake.EliminatedCause, snake.EliminatedBy, snake.EliminatedOnTurn, "snake length", len(snake.Body))
		}
		return
	}
	println("child: {")
	best_node := node.children[0]
	for _, child := range node.children {
		println("move", child.action.Move, "[", child.wins, ",", child.sims, "]")
		if child.sims > best_node.sims {
			best_node = child
		}
	}
	best_node.recur_print()
	print("}")
}

func (node *Node) select_best_move(snake_id string, name string) rules.SnakeMove {

	best_node := node.children[0]

	for _, child := range node.children {
		println(child.action.Move, child.sims, child.wins)

		val := child.sims
		if val > best_node.sims {
			best_node = child
		}
	}

	println(name, "selected best move", best_node.action.Move, "on turn", node.board.board.Turn)

	for _, food := range node.board.board.Food {
		println("food loc (", food.X, food.Y, ")")
	}
	best_node.recur_print()
	return best_node.action
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
	return reward + discover

}

func (node *Node) play_out() {
	iterations := 0
	copy_board := node.board.copy()
	game_over, _ := node.board.rules_set.IsGameOver(&copy_board.board)
	copy_board.rules_set.FoodSpawnChance /= 2
	copy_board.settings.FoodSpawnChance /= 2
	current_turn := node.get_next_player(node.player)

	for !game_over {

		check_game_over, _ := node.board.rules_set.IsGameOver(&copy_board.board)

		if check_game_over {
			break
		}

		moves := copy_board.getValidMoves(current_turn)

		if len(moves) == 0 {
			moves = append(moves, rules.SnakeMove{Move: rules.MoveDown, ID: current_turn})
		}

		selected_move := moves[rand.Intn(len(moves))]
		last_in_rotation := node.player_order[current_turn] == (len(node.player_arr) - 1)
		new_game_over, new_board, err := copy_board.executeAction(selected_move, last_in_rotation)
		copy_board.board = *new_board
		game_over = new_game_over
		current_turn = node.get_next_player(current_turn)

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
	return "tie"
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
	last_in_rotation := parent.player_order[player] == (len(parent.player_arr) - 1)
	_, new_board, _ := board_copy.executeAction(action, last_in_rotation)
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
