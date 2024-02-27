package emu

import (
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/require"
)

func makeLine(text string) Line {
	line := make(Line, 0)

	for _, r := range text {
		glyph := EmptyGlyph()
		glyph.Char = r
		line = append(line, glyph)

		// Handle wider characters
		w := runewidth.RuneWidth(r)
		if w > 1 {
			for i := 0; i < w-1; i++ {
				line = append(line, EmptyGlyph())
			}
		}
	}

	return line
}

func makeWrapped(lines ...string) []Line {
	result := make([]Line, 0)
	for i, str := range lines {
		line := makeLine(str)
		if i != len(lines)-1 {
			line[len(line)-1].Mode |= attrWrap
		}
		result = append(result, line)
	}
	return result
}

func ensureWrap(t *testing.T, input string, cols int, expected []Line) {
	result := wrapLine(makeLine(input), cols)
	require.Equal(t, expected, result)
}

func TestWrap(t *testing.T) {
	ensureWrap(t, "a a", 2, makeWrapped(
		"a ",
		"a ",
	))
	ensureWrap(t, "bbbb", 2, makeWrapped(
		"bb",
		"bb",
	))
	ensureWrap(t, "a   ", 2, makeWrapped(
		"a ",
	))
	ensureWrap(t, "你好", 2, makeWrapped(
		"你",
		"好",
	))
}

func TestLongLine(t *testing.T) {
	term := New()

	for i := 0; i < 40; i++ {
		term.Write([]byte("a"))
	}
	for i := 0; i < 40; i++ {
		term.Write([]byte("b"))
	}
	term.Resize(40, 24)
	require.Equal(t, "a", extractStr(term, 39, 39, 0))
	require.Equal(t, "b", extractStr(term, 0, 0, 1))
	term.Resize(80, 24)
	require.Equal(t, "b", extractStr(term, 40, 40, 0))
}

// Ensure that returning the terminal to its original size puts lines back into
// their original location.
func TestSeveralLines(t *testing.T) {
	term := New()
	term.Resize(4, 24)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("test\ntest"))
	term.Resize(2, 24)
	term.Resize(4, 24)
	require.Equal(t, "test", extractStr(term, 0, 3, 0))
	require.Equal(t, "test", extractStr(term, 0, 3, 1))
}

// Ensure that wrapped lines that continue from the history onto the screen are
// moved completely into history when a resize occurs.
func TestDisappear(t *testing.T) {
	term := New()
	term.Resize(4, 2)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("foobar\ntest"))
	require.Equal(t, "ar", extractStr(term, 0, 1, 0))
	require.Equal(t, "test", extractStr(term, 0, 3, 1))
	term.Resize(4, 3)
	require.Equal(t, "test", extractStr(term, 0, 3, 0))
	history := term.History()
	require.Equal(t, len(history), 1)
	require.Equal(t, "foobar", history[0].String())
}

// A more constrained version of TestSeveralLines.
func TestExpand(t *testing.T) {
	term := New()
	term.Resize(4, 4)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("test\ntest"))
	term.Resize(2, 4)
	require.Equal(t, "te", extractStr(term, 0, 1, 0))
	require.Equal(t, "st", extractStr(term, 0, 1, 1))
	require.Equal(t, "te", extractStr(term, 0, 1, 2))
	require.Equal(t, "st", extractStr(term, 0, 1, 3))
	term.Resize(4, 4)
	require.Equal(t, "test", extractStr(term, 0, 3, 0))
	require.Equal(t, "test", extractStr(term, 0, 3, 1))
	require.Equal(t, "    ", extractStr(term, 0, 3, 2))
}

// Like TestExpand, but even more constrained, which pushes a line into the
// history.
func TestFull(t *testing.T) {
	term := New()
	term.Resize(4, 2)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("test\ntest"))
	require.Equal(t, "test", extractStr(term, 0, 3, 0))
	require.Equal(t, "test", extractStr(term, 0, 3, 1))
	term.Resize(2, 2)
	require.Equal(t, "te", extractStr(term, 0, 1, 0))
	require.Equal(t, "st", extractStr(term, 0, 1, 1))
	term.Resize(4, 2)
	require.Equal(t, "test", extractStr(term, 0, 3, 0))
	require.Equal(t, "    ", extractStr(term, 0, 3, 1))

	history := term.History()
	require.Equal(t, len(history), 1)
	require.Equal(t, "test", history[0].String())
}

// Ensure that when the terminal is on the alt screen, resizing causes a reflow
// of the main screen.
func TestAlt(t *testing.T) {
	term := New()
	term.Resize(4, 4)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("test"))
	term.Write([]byte("\033[?1049h")) // enter altscreen
	term.Write([]byte("foobar foobar foobar"))
	term.Resize(2, 4)
	term.Write([]byte("\033[?1049l")) // leave altscreen
	require.Equal(t, "te", extractStr(term, 0, 1, 0))
	require.Equal(t, "st", extractStr(term, 0, 1, 1))
}
