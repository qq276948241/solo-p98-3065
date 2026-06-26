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

	Render(game.Snake, game.Food, game.Obstacles, game.Score, game.HighScore, game.SpeedLevel, game.IsGameOver(), game.IsPaused())

	for game.Running {
		game.PollInput()

		select {
		case <-game.Ticker.C:
			if game.IsActive() {
				game.Step()
			}
			Render(game.Snake, game.Food, game.Obstacles, game.Score, game.HighScore, game.SpeedLevel, game.IsGameOver(), game.IsPaused())
		default:
		}
	}

	MoveCursor(0, MapHeight+5)
}
