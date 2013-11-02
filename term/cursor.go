package term

import (
	"github.com/nsf/termbox-go"
)

type Cursor struct {
	X, Y        int
	screen      *Screen
	currentLine Line
}

func NewCursor(s *Screen) *Cursor {
	c := &Cursor{
		X:      0,
		Y:      1,
		screen: s,
	}
	c.set()

	return c
}

func (c *Cursor) Hide() {
	termbox.HideCursor()
}

func (c *Cursor) Down() {
	if c.currentLine != nil {
		c.currentLine.Display(0, c.Y, c.screen)
	}
	c.Y++
	l := c.screen.Lines[c.Y-1]
	if err := l.Highlight(0, c.Y, c.screen); err != nil {
		panic(err)
	}
	c.currentLine = l
	c.set()
}

func (c *Cursor) Up() {
	if c.currentLine != nil {
		c.currentLine.Display(0, c.Y, c.screen)
	}
	c.Y--
	l := c.screen.Lines[c.Y-1]
	if err := l.Highlight(0, c.Y, c.screen); err != nil {
		panic(err)
	}
	c.currentLine = l

	c.set()
}

func (c *Cursor) Select() {
	if s, ok := c.currentLine.(Selectable); ok {
		s.Select(c.screen)
	}
}

func (c *Cursor) set() {
	termbox.SetCursor(c.X, c.Y)
	termbox.Flush()
}
