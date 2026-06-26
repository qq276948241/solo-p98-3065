package main

import "math/rand"

type Food struct {
	Pos Point
}

func NewFood() *Food {
	return &Food{}
}

func (f *Food) Spawn(snake *Snake) {
	empty := make([]Point, 0, MapWidth*MapHeight)
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			p := Point{X: x, Y: y}
			if !snake.Occupies(p) {
				empty = append(empty, p)
			}
		}
	}
	if len(empty) > 0 {
		f.Pos = empty[rand.Intn(len(empty))]
	}
}

func (f *Food) EatenBy(head Point) bool {
	return f.Pos.X == head.X && f.Pos.Y == head.Y
}
