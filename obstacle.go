package main

import "math/rand"

type Obstacles []Point

const (
	minObstacles = 3
	maxObstacles = 5
	safetyRadius  = 2
)

func NewObstacles(snake *Snake) Obstacles {
	count := minObstacles + rand.Intn(maxObstacles-minObstacles+1)
	obs := make(Obstacles, 0, count)

	used := make(map[Point]bool)
	for _, seg := range snake.Body {
		used[seg] = true
	}

	for _, p := range snake.nearbyPoints(safetyRadius) {
		used[p] = true
	}

	candidates := make([]Point, 0, MapWidth*MapHeight)
	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			p := Point{X: x, Y: y}
			if !used[p] {
				candidates = append(candidates, p)
			}
		}
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for i := 0; i < count && i < len(candidates); i++ {
		obs = append(obs, candidates[i])
	}

	return obs
}

func (o Obstacles) Has(p Point) bool {
	for _, op := range o {
		if op.X == p.X && op.Y == p.Y {
			return true
		}
	}
	return false
}

func (s *Snake) nearbyPoints(radius int) []Point {
	res := make([]Point, 0)
	for _, seg := range s.Body {
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				nx, ny := seg.X+dx, seg.Y+dy
				if nx >= 0 && nx < MapWidth && ny >= 0 && ny < MapHeight {
					res = append(res, Point{X: nx, Y: ny})
				}
			}
		}
	}
	return res
}
