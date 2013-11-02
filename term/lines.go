package term

import (
	"github.com/nsf/termbox-go"
)

type Line interface {
	Display(x, y int, s *Screen) error
	Highlight(x, y int, s *Screen) error
}

type Selectable interface {
	Select(s *Screen) error
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

func (l *TextLine) Highlight(x, y int, s *Screen) error {
	cb := l.Background
	l.Background = termbox.ColorYellow
	err := l.Display(x, y, s)
	l.Background = cb
	return err
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
func (l *CellLine) Highlight(x, y int, s *Screen) error {
	cb := l.Cells[0].Background

	for _, c := range l.Cells {
		c.Background = termbox.ColorYellow
	}
	err := l.Display(x, y, s)
	for _, c := range l.Cells {
		c.Background = cb
	}

	return err
}

type CellLine struct {
	Cells []*Cell
}

type SelectableLine struct {
	line   Line
	Action func() error
}

func NewSelectableLine(l Line, action func() error) Line {
	return &SelectableLine{
		line:   l,
		Action: action,
	}
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

func (c *SelectableLine) Select(screen *Screen) error {
	return c.Action()
}

func (c *SelectableLine) Display(x, y int, s *Screen) error {
	return c.line.Display(x, y, s)
}

func (c *SelectableLine) Highlight(x, y int, s *Screen) error {
	return c.line.Highlight(x, y, s)
}

func writeSpacer(x, y int, s *Screen) int {
	for i := 0; i < 2; i++ {
		termbox.SetCell(x, y, ' ', s.DefaultForground, s.DefaultBackground)
		x++
	}
	return 2
}
