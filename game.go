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
		if !g.GameOver {
			g.Paused = !g.Paused
		}
	case ActionDir:
		if !g.GameOver && !g.Paused {
			g.Snake.SetDirection(e.Direction)
		}
	}
}

func (g *Game) Step() {
	if g.GameOver || g.Paused {
		return
	}

	newHead := g.Snake.Move()

	if g.Snake.CollidesWall(newHead) {
		g.endGame()
		return
	}
	if g.Snake.CollidesSelf(newHead) {
		g.endGame()
		return
	}
	if g.Obstacles.Has(newHead) {
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
