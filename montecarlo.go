package main

import (
	"math"
	"math/rand"

	"github.com/BattlesnakeOfficial/rules"
)

type Tree struct {
	player string
	root   *Node
}

type Node struct {
	children   []*Node
	parent     *Node
	board      Simulation
	sims       int
	table      map[string]int
	joint_move []rules.SnakeMove
}

const c float64 = 2.141

func new_tree(my_snake string, game GameState) Tree {
	return Tree{
		player: my_snake,
		root: &Node{
			children: []*Node{},
			parent:   nil,
			board:    simulationFromGame(&game),
			table:    make(map[string]int),
		},
	}
}

func (tree *Tree) monte_move() rules.SnakeMove {

	tree.root.expandNode()

	for i := 0; i < 1000; i++ {
		tree.expand_tree()
	}

	return tree.root.select_best_move(tree.player)
}

func (node *Node) select_best_move(snake_id string) rules.SnakeMove {

	sims_for_move := make(map[string]int)

	for _, child := range node.children {
		move := get_move_by_snake(snake_id, child.joint_move)
		add_to_map(sims_for_move, move.Move, child.sims)
	}

	most_sims := 0
	best_move := rules.MoveUp

	for move, sims := range sims_for_move {
		if sims > most_sims {
			most_sims = sims
			best_move = move
		}
	}
	return rules.SnakeMove{ID: snake_id, Move: best_move}
}

func (tree *Tree) expand_tree() {
	var promising_node = tree.root.select_node()

	promising_node.expandNode()

	var test_node = promising_node

	if len(promising_node.children) > 0 {
		test_node = promising_node.children[(int)(math.Floor((float64)(rand.Intn(len(promising_node.children)))))]
	}

	test_node.play_out()
}

func (node *Node) expandNode() {

	move_matrix := node.board.generateMoveMatrix()

	for _, joint_move := range move_matrix {
		node.children = append(node.children, create_child(node, joint_move, node.board))
	}

}

func (node *Node) select_node() *Node {
	var selected_moves = []rules.SnakeMove{}
	if len(node.children) > 0 {
		for _, snake := range node.board.board.Snakes {
			move_win_sum := make(map[string]int)
			move_sim_sum := make(map[string]int)
			for _, child := range node.children {
				move := get_move_by_snake(snake.ID, child.joint_move)
				add_to_map(move_sim_sum, move.Move, child.sims)
				add_to_map(move_win_sum, move.Move, child.table[snake.ID])
			}

			// get max move for the snake
			best_move := rules.MoveUp
			var best_val float64 = 0

			for move, sims := range move_sim_sum {

				utc_val := calc_utc_val(move_win_sum[move], sims, node.parent.sims)

				if utc_val > float64(best_val) {
					best_move = move
					best_val = utc_val
				}
			}

			selected_moves = append(selected_moves, rules.SnakeMove{
				ID:   snake.ID,
				Move: best_move,
			})
		}

		for _, child := range node.children {
			if compare_joint_move(child.joint_move, selected_moves) {
				return child.select_node()
			}
		}

	}
	return node
}

func compare_joint_move(joint_move []rules.SnakeMove, other []rules.SnakeMove) bool {
	matches := 0

	for _, move := range joint_move {
		for _, move2 := range other {
			if move.ID == move2.ID && move2.Move == move.Move {
				matches += 1
			}
		}
	}

	return matches == len(joint_move)
}

func calc_utc_val(wins int, sims int, parent_sims int) float64 {
	return ((float64)(wins) / (float64)(sims)) + (c * math.Pow(math.Log((float64)(parent_sims))/(float64)(sims), .5))

}

func add_to_map[K comparable](m map[K]int, key K, val int) {
	val, exists := m[key]
	if exists {
		m[key] += val
	} else {
		m[key] = val
	}
}

func get_move_by_snake(snake_id string, joint_move []rules.SnakeMove) rules.SnakeMove {
	for _, move := range joint_move {
		if move.ID == snake_id {
			return move
		}
	}
	panic("for some reason there was no move found")
}

func (node *Node) play_out() {
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

		if err == nil {
			println(err)
			panic("error thrown while playing out")
		}
	}
	winner := copy_board.board.Snakes[0].ID
	node.back_prop(winner)
}

func (node *Node) back_prop(winner string) {
	add_to_map(node.table, winner, 1)
	node.sims++

	if node.parent != nil {
		node.parent.back_prop(winner)
	}
}

func create_child(parent *Node, joint_move []rules.SnakeMove, board Simulation) *Node {

	board_copy := board.copy()
	board_copy.executeActions(joint_move)

	uct_table := make(map[string]int)

	for _, snake := range board.board.Snakes {
		uct_table[snake.ID] = 0
	}

	return &Node{
		sims:       0,
		table:      uct_table,
		joint_move: joint_move,
		children:   []*Node{},
		parent:     parent,
		board:      board_copy,
	}
}
