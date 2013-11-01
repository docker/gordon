package term

import (
	"github.com/nsf/termbox-go"
)

type Line interface {
	Display(x, y int, s *Screen) error
}

type TextLine struct {
	Content string
}

func (l *TextLine) Display(x, y int, s *Screen) error {
	for _, r := range l.Content {
		termbox.SetCell(x, y, r, s.DefaultForground, s.DefaultBackground)
		y++
	}
	return nil
}
