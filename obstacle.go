package main

import "math/rand"

type Obstacles struct {
	points []Point
	count  int
}

const (
	obstacleMinCount  = 3
	obstacleMaxCount  = 5
	obstacleSafetyRad = 2
)

func NewObstacles(snake *Snake) Obstacles {
	count := obstacleMinCount + rand.Intn(obstacleMaxCount-obstacleMinCount+1)
	obs := Obstacles{
		points: make([]Point, 0, count),
		count:  count,
	}
	obs.generate(snake)
	return obs
}

func (o *Obstacles) generate(snake *Snake) {
	used := make(map[Point]bool)
	for _, seg := range snake.Body {
		used[seg] = true
	}
	for _, p := range nearbyPoints(snake.Body, obstacleSafetyRad) {
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

	for i := 0; i < o.count && i < len(candidates); i++ {
		o.points = append(o.points, candidates[i])
	}
}

func (o Obstacles) Has(p Point) bool {
	for _, op := range o.points {
		if op.X == p.X && op.Y == p.Y {
			return true
		}
	}
	return false
}

func (o Obstacles) List() []Point {
	res := make([]Point, len(o.points))
	copy(res, o.points)
	return res
}

func (o Obstacles) Count() int {
	return len(o.points)
}

func nearbyPoints(body []Point, radius int) []Point {
	res := make([]Point, 0)
	for _, seg := range body {
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
