package main

import "github.com/BattlesnakeOfficial/rules"

type Direction string

type Simulation struct {
	board     rules.BoardState
	settings  rules.Settings
	rules_set rules.StandardRuleset
}

func (sim *Simulation) copy() Simulation {
	return Simulation{
		board:     *sim.board.Clone(),
		settings:  sim.settings,
		rules_set: sim.rules_set,
	}
}

func simulationFromGame(game *GameState) Simulation {
	return Simulation{
		board: rules.BoardState{
			Height: game.Board.Height,
			Width:  game.Board.Width,
			Snakes: convertSnakes(game.Board.Snakes),
		},
	}
}

func (sim *Simulation) generateMoveMatrix() [][]rules.SnakeMove {
	var move_matrix = [][]rules.SnakeMove{}
	for _, snake := range sim.board.Snakes {
		snake_id := snake.ID
		moves := sim.getValidMoves(snake_id)

		if len(moves) == 0 {
			moves = []rules.SnakeMove{{ID: snake_id, Move: rules.MoveUp}}
		}
		var new_matrix = [][]rules.SnakeMove{}

		if len(move_matrix) == 0 {
			for _, move := range moves {
				new_matrix = append(new_matrix, []rules.SnakeMove{move})
			}
		} else {
			for _, move := range moves {
				for _, move_arr := range move_matrix {
					move_arr = append(move_arr, move)
					new_matrix = append(new_matrix, move_arr)
				}
			}
		}

		move_matrix = new_matrix
	}
	return move_matrix
}

func convertSnakes(api_snakes []Battlesnake) []rules.Snake {
	var real_snakes = []rules.Snake{}
	for _, snake := range api_snakes {
		real_snakes = append(real_snakes,
			rules.Snake{
				ID:               snake.ID,
				Body:             convertBody(snake.Body),
				Health:           int(snake.Health),
				EliminatedCause:  rules.NotEliminated,
				EliminatedOnTurn: 0,
				EliminatedBy:     rules.NotEliminated,
			})
	}
	return real_snakes
}

func convertBody(body []Coord) []rules.Point {
	var new_body = []rules.Point{}
	for _, bod := range body {
		new_body = append(new_body, rules.Point{
			X: bod.X,
			Y: bod.Y,
		})
	}
	return new_body
}

func (game *Simulation) getValidMoves(snakeId string) []rules.SnakeMove {

	snake := get_snake(game.board, snakeId)

	var dirs = [][]int{{1, 0}, {-1, 0}, {0, -1}, {0, 1}}

	var valid_moves = []rules.SnakeMove{}

	for _, dir := range dirs {

		x_dir := dir[0]
		y_dir := dir[1]

		head := snake.Body[0]

		new_head := rules.Point{}
		new_head.X = head.X + x_dir
		new_head.Y = head.Y + y_dir

		//check for neck collision
		if len(snake.Body) > 1 {
			neck := snake.Body[1]
			if neck.X == new_head.X && neck.Y == new_head.Y {
				continue
			}
		}

		// check for wall collisions

		if new_head.X >= game.board.Width || new_head.X < 0 || new_head.Y >= game.board.Height || new_head.Y < 0 {
			continue
		}

		valid_move := true

		// check for snake collisions

		for _, other_snake := range game.board.Snakes {

			// don't collide with self
			if other_snake.ID == snakeId {
				continue
			}

			if snakeHasLostHeadToHead(snake, &other_snake) {
				valid_move = false
				break
			}

			if snakeHasBodyCollided(snake, &other_snake) {
				valid_move = false
				break
			}

		}

		if valid_move {
			valid_moves = append(valid_moves, rules.SnakeMove{
				ID:   snakeId,
				Move: getMoveFromDir(dir),
			})
		}
	}

	return valid_moves
}

func getMoveFromDir(dir []int) string {
	if dir[0] == 1 {
		return rules.MoveRight
	}
	if dir[0] == -1 {
		return rules.MoveLeft
	}

	if dir[1] == 1 {
		return rules.MoveUp
	} else {

		return rules.MoveDown
	}
}

func snakeHasLostHeadToHead(s *rules.Snake, other *rules.Snake) bool {
	if s.Body[0].X == other.Body[0].X && s.Body[0].Y == other.Body[0].Y {
		return len(s.Body) <= len(other.Body)
	}
	return false
}

func snakeHasBodyCollided(s *rules.Snake, other *rules.Snake) bool {
	head := s.Body[0]
	for i, body := range other.Body {
		if i == 0 {
			continue
		} else if head.X == body.X && head.Y == body.Y {
			return true
		}
	}
	return false
}

func get_snake(board rules.BoardState, snakeId string) *rules.Snake {
	for _, snake := range board.Snakes {
		if snake.ID == snakeId {
			return &snake
		}
	}
	return nil
}

func (game *Simulation) executeActions(moves []rules.SnakeMove) (bool, *rules.BoardState, error) {
	return game.rules_set.Execute(&game.board, game.settings, moves)
}
