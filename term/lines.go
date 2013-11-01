package term

import (
	"github.com/nsf/termbox-go"
)

type Line interface {
	Display(x, y int, s *Screen) error
}

type TextLine struct {
	Content               string
	Forground, Background termbox.Attribute
}

func (l *TextLine) Display(x, y int, s *Screen) error {
	for _, r := range l.Content {
		termbox.SetCell(x, y, r, l.Forground, l.Background)
		x++
	}
	for i := x; i < s.Width; i++ {
		termbox.SetCell(i, y, ' ', l.Forground, l.Background)
	}
	return nil
}

type Cell struct {
	Content               string
	Forground, Background termbox.Attribute
	Width                 int
}

func NewCell(content string, fg, bg termbox.Attribute) *Cell {
	return &Cell{
		Content:    content,
		Width:      len(content),
		Forground:  fg,
		Background: bg,
	}
}

func (c *Cell) Display(x, y int) error {
	cCount := 0
	for _, r := range c.Content {
		if cCount >= c.Width {
			break
		}
		termbox.SetCell(x, y, r, c.Forground, c.Background)
		x++
		cCount++
	}
	for i := x; i < c.Width; i++ {

		termbox.SetCell(i, y, ' ', c.Forground, c.Background)
	}
	return nil
}

type CellLine struct {
	Cells []*Cell
}

func (l *CellLine) Display(x, y int, s *Screen) error {
	for _, c := range l.Cells {
		if err := c.Display(x, y); err != nil {
			return err
		}
		x = x + c.Width
		x = x + writeSpacer(x, y, s)
	}
	return nil
}

func writeSpacer(x, y int, s *Screen) int {
	for i := 0; i < 2; i++ {
		termbox.SetCell(x, y, ' ', s.DefaultForground, s.DefaultBackground)
		x++
	}
	return 2
}
