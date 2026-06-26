package main

func main() {
	InitKeyboard()
	defer RestoreKeyboard()
	InitRender()

	ClearScreen()

	game := NewGame()
	defer func() {
		if game.Ticker != nil {
			game.Ticker.Stop()
		}
	}()

	Render(game.Snake, game.Food, game.Obstacles, game.Score, game.HighScore, game.SpeedLevel, game.GameOver, game.Paused)

	for game.Running {
		game.PollInput()

		select {
		case <-game.Ticker.C:
			if !game.GameOver && !game.Paused {
				game.Step()
			}
			Render(game.Snake, game.Food, game.Obstacles, game.Score, game.HighScore, game.SpeedLevel, game.GameOver, game.Paused)
		default:
		}
	}

	MoveCursor(0, MapHeight+5)
}
