package main

import (
	"github.com/BattlesnakeOfficial/rules"
)

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
			Turn:   game.Turn,
			Height: game.Board.Height,
			Width:  game.Board.Width,
			Snakes: convertSnakes(game.Board.Snakes),
			Food:   convert_food(game.Board.Food),
		},
		rules_set: convert_ruleset(game.Game.Ruleset),
		settings:  convert_settings(game.Game.Ruleset.Settings),
	}
}

func convert_food(food []Coord) []rules.Point {
	real_food := []rules.Point{}
	for _, f := range food {
		real_food = append(real_food, rules.Point{
			X: f.X,
			Y: f.Y,
		})
	}
	return real_food
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
			moves = []rules.SnakeMove{{ID: snake_id, Move: rules.MoveDown}}
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

func move_point(point rules.Point, appliedMove string) rules.Point {

	new_point := rules.Point{}

	switch appliedMove {
	// Guaranteed to be one of these options given the clause above
	case rules.MoveUp:
		new_point.X = int(point.X)
		new_point.Y = int(point.Y + 1)
	case rules.MoveDown:
		new_point.X = int(point.X)
		new_point.Y = int(point.Y - 1)
	case rules.MoveLeft:
		new_point.X = int(point.X - 1)
		new_point.Y = int(point.Y)
	case rules.MoveRight:
		new_point.X = int(point.X + 1)
		new_point.Y = int(point.Y)
	}
	return new_point
}

func copy_point(point rules.Point) rules.Point {
	return rules.Point{
		X: int(point.X),
		Y: int(point.Y),
	}
}

func copy_snake(snake rules.Snake) *rules.Snake {
	new_body := make([]rules.Point, len(snake.Body))

	for _, bod := range snake.Body {
		new_body = append(new_body, copy_point(bod))
	}

	return &rules.Snake{
		ID:   snake.ID,
		Body: new_body,
	}
}

func snakeIsOutOfBounds(s *rules.Snake, boardWidth int, boardHeight int) bool {
	point := s.Body[0]
	if (point.X < 0) || (point.X >= boardWidth) {
		return true
	}
	if (point.Y < 0) || (point.Y >= boardHeight) {
		return true
	}

	return false
}

func (game *Simulation) getValidMoves(snakeId string) []rules.SnakeMove {

	snake := get_snake(game.board, snakeId)

	var dirs = []string{rules.MoveUp, rules.MoveDown, rules.MoveLeft, rules.MoveRight}

	var valid_moves = []rules.SnakeMove{}

	if snake.EliminatedCause != "" {
		return valid_moves
	}

	for _, dir := range dirs {

		snake_moved := move_point(snake.Body[0], dir)

		// check for wall collisions
		if snake_moved.X >= game.board.Width || snake_moved.X < 0 || snake_moved.Y >= game.board.Height || snake_moved.Y < 0 {
			// println("avoided wall collision", snake_moved.X, snake_moved.Y)
			continue
		}

		valid := true

		for _, bod := range snake.Body {
			if snake_moved.X == bod.X && snake_moved.Y == bod.Y {
				valid = false
				break
			}
		}

		// for _, snake := range game.board.Snakes {
		// 	// if snake_self_collided(&snake_moved, &snake) {
		// 	// 	valid = false
		// 	// 	break
		// 	// }

		// 	if snake.ID == snake_moved.ID {
		// 		continue
		// 	}

		// 	if snakeHasLostHeadToHead(&snake_moved, &snake) {
		// 		valid = false
		// 		break
		// 	}
		// }

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

func (game *Simulation) executeAction(move rules.SnakeMove) (bool, *rules.BoardState, error) {
	move_arr := []rules.SnakeMove{move}

	// StageGameOverStandard,
	// StageMovementStandard,
	// StageStarvationStandard,
	// StageHazardDamageStandard,
	// StageFeedSnakesStandard,
	// StageEliminationStandard,

	_, err1 := MoveSnakesStandard(&game.board, game.settings, move_arr)
	if err1 != nil {
		panic(err1.Error())

	}
	_, err4 := rules.EliminateSnakesStandard(&game.board, game.settings, move_arr)
	if err4 != nil {
		panic(err4.Error())
	}
	_, err2 := rules.ReduceSnakeHealthStandard(&game.board, game.settings, move_arr)
	if err2 != nil {
		panic(err2.Error())
	}
	_, err3 := rules.FeedSnakesStandard(&game.board, game.settings, move_arr)
	if err3 != nil {
		panic(err3.Error())
	}

	game_over, err := rules.GameOverStandard(&game.board, game.settings, move_arr)

	if game_over {
		return game_over, &game.board, err
	}
	return game_over, &game.board, nil
}

func MoveSnakesStandard(b *rules.BoardState, settings rules.Settings, moves []rules.SnakeMove) (bool, error) {
	if rules.IsInitialization(b, settings, moves) {
		return false, nil
	}

	// no-op when moves are empty
	if len(moves) == 0 {
		return false, nil
	}

	// Sanity check that all non-eliminated snakes have moves and bodies.
	for i := 0; i < len(b.Snakes); i++ {
		snake := &b.Snakes[i]
		if snake.EliminatedCause != rules.NotEliminated {
			continue
		}

		if len(snake.Body) == 0 {
			return false, rules.ErrorZeroLengthSnake
		}

	}

	for i := 0; i < len(b.Snakes); i++ {
		snake := &b.Snakes[i]
		if snake.EliminatedCause != rules.NotEliminated {
			continue
		}

		for _, move := range moves {
			if move.ID == snake.ID {
				appliedMove := move.Move
				switch move.Move {
				case rules.MoveUp, rules.MoveDown, rules.MoveRight, rules.MoveLeft:
					break
				}

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

				// Append new head, pop old tail
				snake.Body = append([]rules.Point{newHead}, snake.Body[:len(snake.Body)-1]...)
			}
		}
	}
	return false, nil
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
