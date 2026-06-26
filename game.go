package main

import (
	"math/rand"
	"time"
)

type Game struct {
	Snake      *Snake
	Food       *Food
	Obstacles  Obstacles
	Score      int
	HighScore  int
	SpeedLevel int
	GameOver   bool
	Paused     bool
	Ticker     *time.Ticker
	Running    bool
}

const (
	baseTick     = 200 * time.Millisecond
	speedStep    = 50
	minTick      = 50 * time.Millisecond
	tickDecrease = 25 * time.Millisecond
)

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	snake := NewSnake()
	obstacles := NewObstacles(snake)
	food := NewFood()
	food.Spawn(snake, obstacles)
	highScore := LoadHighScore()
	game := &Game{
		Snake:      snake,
		Food:       food,
		Obstacles:  obstacles,
		Score:      0,
		HighScore:  highScore,
		SpeedLevel: 1,
		GameOver:   false,
		Paused:     false,
		Running:    true,
	}
	game.resetTicker()
	return game
}

func (g *Game) IsGameOver() bool {
	return g.GameOver
}

func (g *Game) IsPaused() bool {
	return g.Paused
}

func (g *Game) IsActive() bool {
	return !g.GameOver && !g.Paused
}

func (g *Game) CanChangeDirection() bool {
	return !g.GameOver && !g.Paused
}

func (g *Game) CanTogglePause() bool {
	return !g.GameOver
}

func (g *Game) TogglePause() {
	if g.CanTogglePause() {
		g.Paused = !g.Paused
	}
}

func (g *Game) resetTicker() {
	if g.Ticker != nil {
		g.Ticker.Stop()
	}
	tick := baseTick - time.Duration(g.SpeedLevel-1)*tickDecrease
	if tick < minTick {
		tick = minTick
	}
	g.Ticker = time.NewTicker(tick)
}

func (g *Game) Restart() {
	rand.Seed(time.Now().UnixNano())
	g.Snake = NewSnake()
	g.Obstacles = NewObstacles(g.Snake)
	g.Food = NewFood()
	g.Food.Spawn(g.Snake, g.Obstacles)
	g.Score = 0
	g.SpeedLevel = 1
	g.GameOver = false
	g.Paused = false
	g.resetTicker()
}

func (g *Game) updateSpeed() {
	desiredLevel := g.Score/speedStep + 1
	if desiredLevel > g.SpeedLevel {
		g.SpeedLevel = desiredLevel
		g.resetTicker()
	}
}

func (g *Game) HandleKey(e KeyEvent) {
	switch e.Action {
	case ActionQuit:
		g.Running = false
	case ActionRestart:
		g.Restart()
	case ActionPause:
		g.TogglePause()
	case ActionDir:
		if g.CanChangeDirection() {
			g.Snake.SetDirection(e.Direction)
		}
	}
}

func (g *Game) Step() {
	if !g.IsActive() {
		return
	}

	newHead := g.Snake.Move()

	if g.checkCollision(newHead) {
		g.endGame()
		return
	}

	ateFood := g.Food.EatenBy(newHead)
	g.Snake.Advance(newHead, ateFood)

	if ateFood {
		g.Score += 10
		g.Food.Spawn(g.Snake, g.Obstacles)
		g.updateSpeed()
	}
}

func (g *Game) checkCollision(p Point) bool {
	return g.Snake.CollidesWall(p) || g.Snake.CollidesSelf(p) || g.Obstacles.Has(p)
}

func (g *Game) endGame() {
	g.GameOver = true
	g.HighScore = UpdateHighScoreIfNeeded(g.Score, g.HighScore)
	g.Ticker.Stop()
}

func (g *Game) PollInput() {
	for {
		e := ReadKeyEvent()
		if e.Action == ActionNone {
			return
		}
		g.HandleKey(e)
	}
}

func (g *Game) GetCellContent(x, y int) string {
	p := Point{X: x, Y: y}
	if g.Snake.IsHead(p) {
		return "@"
	}
	if g.Snake.Occupies(p) {
		return "o"
	}
	if g.Food.Pos.X == x && g.Food.Pos.Y == y {
		return "*"
	}
	if g.Obstacles.Has(p) {
		return "#"
	}
	return " "
}
