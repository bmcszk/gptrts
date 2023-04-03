package game

import (
	"image"
	"math"
)

type PF struct {
	X float64
	Y float64
}

func NewPF(x, y float64) PF {
	return PF{
		X: x, Y: y,
	}
}

func (p PF) ImagePoint() image.Point {
	return image.Pt(int(p.X), int(p.Y))
}

func (p PF) Add(p2 PF) PF {
	return NewPF(p.X+p2.X, p.Y+p2.Y)
}

func (p PF) Mul(a float64) PF {
	return NewPF(p.X*a, p.Y*a)
}

func (p PF) Step(target PF) PF {
	s := p.Round()
	target = target.Round()
	dx, dy := target.X-s.X, target.Y-s.Y
	if dx > 0 {
		dx = 1
	} else if dx < 0 {
		dx = -1
	}
	if dy > 0 {
		dy = 1
	} else if dy < 0 {
		dy = -1
	}
	return NewPF(s.X+dx, s.Y+dy)
}

func (p PF) Round() PF {
	return NewPF(math.Round(p.X), math.Round(p.Y))
}

func (p PF) Ints() (int, int) {
	p = p.Round()
	return int(p.X), int(p.Y)
}
