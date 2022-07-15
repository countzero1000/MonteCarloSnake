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
		rules_set: convert_ruleset(game.Game.Ruleset),
		settings:  convert_settings(game.Game.Ruleset.Settings),
	}
}

func convert_settings(settings Settings) rules.Settings {
	return rules.Settings{
		FoodSpawnChance:     int(settings.FoodSpawnChance),
		MinimumFood:         int(settings.MinimumFood),
		HazardDamagePerTurn: int(settings.HazardDamagePerTurn),
		RoyaleSettings: rules.RoyaleSettings{
			ShrinkEveryNTurns: int(settings.Royale.ShrinkEveryNTurns),
		},
		SquadSettings: rules.SquadSettings(settings.Squad),
	}
}

func convert_ruleset(ruleset Ruleset) rules.StandardRuleset {
	return rules.StandardRuleset{
		FoodSpawnChance:     int(ruleset.Settings.FoodSpawnChance),
		MinimumFood:         int(ruleset.Settings.MinimumFood),
		HazardDamagePerTurn: int(ruleset.Settings.HazardDamagePerTurn),
	}
}

func (sim *Simulation) generateMoveMatrix() [][]rules.SnakeMove {
	var move_matrix = [][]rules.SnakeMove{}
	for _, snake := range sim.board.Snakes {
		snake_id := snake.ID
		moves := sim.getValidMoves(snake_id)

		if len(moves) == 0 {
			continue
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

func move_snake(snake *rules.Snake, appliedMove string) rules.Snake {
	newHead := rules.Point{}
	switch appliedMove {
	// Guaranteed to be one of these options given the clause above
	case rules.MoveUp:
		newHead.X = snake.Body[0].X
		newHead.Y = snake.Body[0].Y + 1
	case rules.MoveDown:
		newHead.X = snake.Body[0].X
		newHead.Y = snake.Body[0].Y - 1
	case rules.MoveLeft:
		newHead.X = snake.Body[0].X - 1
		newHead.Y = snake.Body[0].Y
	case rules.MoveRight:
		newHead.X = snake.Body[0].X + 1
		newHead.Y = snake.Body[0].Y
	}
	snake.Body = append([]rules.Point{newHead}, snake.Body[:len(snake.Body)-1]...)
	return *snake
}

func copy_snake(snake rules.Snake) *rules.Snake {
	new_body := []rules.Point{}

	for _, p := range snake.Body {
		new_body = append(new_body, p)
	}

	return &rules.Snake{
		ID:   snake.ID,
		Body: new_body,
	}
}

func (game *Simulation) getValidMoves(snakeId string) []rules.SnakeMove {

	snake := get_snake(game.board, snakeId)

	var dirs = []string{rules.MoveUp, rules.MoveDown, rules.MoveLeft, rules.MoveRight}

	var valid_moves = []rules.SnakeMove{}

	if snake.EliminatedCause != "" {
		return valid_moves
	}

	for _, dir := range dirs {

		snake_moved := move_snake(copy_snake(*snake), dir)
		new_head := snake_moved.Body[0]

		// check for wall collisions
		if new_head.X >= game.board.Width || new_head.X < 0 || new_head.Y >= game.board.Height || new_head.Y < 0 {
			continue
		}

		valid := true

		for _, snake := range game.board.Snakes {
			if snake_self_collided(&snake_moved, &snake) {
				valid = false
				break
			}
			if snake.ID == snake_moved.ID {
				continue
			}

			if snakeHasLostHeadToHead(&snake_moved, &snake) {
				valid = false
				break
			}
		}

		if valid {
			valid_moves = append(valid_moves, rules.SnakeMove{
				ID:   snakeId,
				Move: dir,
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

func snake_self_collided(s *rules.Snake, other *rules.Snake) bool {
	head := s.Body[0]
	for i, body := range other.Body {
		if i == 0 {
			continue
		}
		if head.X == body.X && head.Y == body.Y {
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
	if len(moves) < len(game.board.Snakes) {
		missing := find_missing_snakes(moves, game.board.Snakes)
		for _, mimissing_id := range missing {
			moves = append(moves, rules.SnakeMove{
				ID:   mimissing_id,
				Move: rules.MoveDown,
			})
		}
	}
	return game.rules_set.Execute(&game.board, game.settings, moves)
}

func find_missing_snakes(moves []rules.SnakeMove, snakes []rules.Snake) []string {
	missing := []string{}
	for _, snake := range snakes {
		found := false
		for _, move := range moves {
			if snake.ID == move.ID {
				found = true
			}
		}
		if !found {
			missing = append(missing, snake.ID)
		}
	}
	return missing
}
