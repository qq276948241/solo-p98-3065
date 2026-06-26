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

	if game.NeedRender {
		Render(game.Snake, game.Food, game.Obstacles, game.Score, game.HighScore, game.SpeedLevel, game.IsGameOver(), game.IsPaused())
		game.NeedRender = false
	}

	for game.Running {
		game.PollInput()

		select {
		case <-game.Ticker.C:
			if game.IsActive() {
				game.Step()
			} else {
				game.MarkDirty()
			}
		default:
		}

		if game.NeedRender {
			Render(game.Snake, game.Food, game.Obstacles, game.Score, game.HighScore, game.SpeedLevel, game.IsGameOver(), game.IsPaused())
			game.NeedRender = false
		}
	}

	MoveCursor(0, MapHeight+5)
}
