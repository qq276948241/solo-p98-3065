package main

type Point struct {
	X int
	Y int
}

type Snake struct {
	Body      []Point
	Direction Direction
}

const (
	MapWidth  = 20
	MapHeight = 20
)

func NewSnake() *Snake {
	centerX := MapWidth / 2
	centerY := MapHeight / 2
	return &Snake{
		Body: []Point{
			{X: centerX, Y: centerY},
			{X: centerX - 1, Y: centerY},
			{X: centerX - 2, Y: centerY},
		},
		Direction: DirRight,
	}
}

func (s *Snake) Head() Point {
	return s.Body[0]
}

func (s *Snake) IsHead(p Point) bool {
	return s.Body[0].X == p.X && s.Body[0].Y == p.Y
}

func (s *Snake) SetDirection(dir Direction) {
	if (s.Direction == DirUp && dir == DirDown) ||
		(s.Direction == DirDown && dir == DirUp) ||
		(s.Direction == DirLeft && dir == DirRight) ||
		(s.Direction == DirRight && dir == DirLeft) {
		return
	}
	s.Direction = dir
}

func (s *Snake) Move() Point {
	head := s.Head()
	newHead := Point{X: head.X, Y: head.Y}
	switch s.Direction {
	case DirUp:
		newHead.Y--
	case DirDown:
		newHead.Y++
	case DirLeft:
		newHead.X--
	case DirRight:
		newHead.X++
	}
	return newHead
}

func (s *Snake) Advance(newHead Point, grow bool) {
	s.Body = append([]Point{newHead}, s.Body...)
	if !grow {
		s.Body = s.Body[:len(s.Body)-1]
	}
}

func (s *Snake) CollidesWall(p Point) bool {
	return p.X < 0 || p.X >= MapWidth || p.Y < 0 || p.Y >= MapHeight
}

func (s *Snake) CollidesSelf(p Point) bool {
	for _, seg := range s.Body {
		if seg.X == p.X && seg.Y == p.Y {
			return true
		}
	}
	return false
}

func (s *Snake) Occupies(p Point) bool {
	for _, seg := range s.Body {
		if seg.X == p.X && seg.Y == p.Y {
			return true
		}
	}
	return false
}
