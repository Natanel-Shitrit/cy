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
	line := makeLine(input)
	lines := []Line{}
	for _, r := range wrapLine(line, cols) {
		lines = append(
			lines,
			resolveLine([]Line{line}, []ScreenLine{r}),
		)
	}

	for _, line := range expected {
		for i := 0; i < len(line); i++ {
			line[i].Mode = 0
		}
	}

	require.Equal(t, expected, lines)
}

func TestWrap(t *testing.T) {
	ensureWrap(t, "a a", 2, makeWrapped(
		"a ",
		"a",
	))
	ensureWrap(t, "bbbb", 2, makeWrapped(
		"bb",
		"bb",
	))
	ensureWrap(t, "a   ", 2, makeWrapped(
		"a ",
		"  ",
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
	term.Write([]byte("testt"))
	require.Equal(t, 1, term.Cursor().X)
	require.Equal(t, 1, term.Cursor().Y)
	term.Write([]byte(EnterAltScreen))
	term.Write([]byte("foobar foobar foobar"))
	term.Resize(2, 4)
	term.Write([]byte(ExitAltScreen)) // leave altscreen
	require.Equal(t, "te", extractStr(term, 0, 1, 0))
	require.Equal(t, "st", extractStr(term, 0, 1, 1))
	require.Equal(t, 1, term.Cursor().X)
	require.Equal(t, 2, term.Cursor().Y)
}

// Ensure that the cursor remains stationary relative to the physical line it's
// on.
func TestCursor(t *testing.T) {
	term := New()
	term.Resize(6, 4)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("foobar\ntest"))
	require.Equal(t, 4, term.Cursor().X)
	require.Equal(t, 1, term.Cursor().Y)
	term.Resize(5, 4)
	require.Equal(t, 4, term.Cursor().X)
	require.Equal(t, 2, term.Cursor().Y)

	// There was a bug where our cursor was being cleared incorrectly --
	// check for it
	screen := term.Screen()
	cell := screen[len(screen)-1][0]
	require.Equal(t, DefaultFG, cell.FG)
	require.Equal(t, DefaultBG, cell.BG)
}

// Ensure that the (full) main screen is reflowed correctly when on the alt
// screen.
func TestFullAlt(t *testing.T) {
	term := New()
	term.Resize(4, 3)
	term.Write([]byte(LineFeedMode))
	term.Write([]byte("test\ntest\nte"))
	require.Equal(t, 2, term.Cursor().X)
	require.Equal(t, 2, term.Cursor().Y)
	term.Write([]byte(EnterAltScreen))
	term.Resize(2, 3)
	term.Write([]byte(ExitAltScreen))
	require.Equal(t, "te", extractStr(term, 0, 1, 0))
	require.Equal(t, "st", extractStr(term, 0, 1, 1))
	require.Equal(t, "te", extractStr(term, 0, 1, 2))
	require.Equal(t, 1, term.Cursor().X)
	require.Equal(t, 2, term.Cursor().Y)
	require.True(t, term.Cursor().State&cursorWrapNext != 0)
	term.Write([]byte(EnterAltScreen))
	term.Resize(4, 3)
	term.Write([]byte(ExitAltScreen))
	require.Equal(t, "test", extractStr(term, 0, 3, 0))
	require.Equal(t, "te  ", extractStr(term, 0, 3, 1))
	require.Equal(t, "    ", extractStr(term, 0, 3, 2))
	require.Equal(t, 2, term.Cursor().X)
	require.Equal(t, 1, term.Cursor().Y)

	// fill up the rest
	term.Write([]byte("st\ntest\nok"))
	term.Write([]byte(EnterAltScreen))
	term.Resize(2, 5)
	term.Resize(1, 1)
	term.Resize(6, 3)
	term.Write([]byte(ExitAltScreen))
}

// Ensure that terminals that disable history and those that do not end up with
// the same results.
func TestHistory(t *testing.T) {
	a := New()
	b := New(WithoutHistory)

	for _, term := range []Terminal{a, b} {
		term.Resize(4, 3)
		term.Write([]byte(LineFeedMode))
		term.Write([]byte("foobar\nfoo\ntest"))
		term.Resize(4, 2)
	}

	require.Equal(t, a.String(), b.String())
}

func translateTest(
	t *testing.T,
	oldCols, newCols int,
	oldCursor, expectedCursor Cursor,
	screen ...string,
) {
	lines := make([]Line, 0)
	for _, line := range screen {
		lines = append(lines, makeLine(line))
	}
	oldPhysical := wrapLines(lines, oldCols)
	newPhysical := wrapLines(lines, newCols)

	newCursor := translateCursor(
		lines, lines,
		oldPhysical, newPhysical,
		oldCursor,
		newCols,
	)

	require.Equal(t, expectedCursor, newCursor.cursor)
}

func TestTranslateCursor(t *testing.T) {
	// Stays at end
	translateTest(t, 4, 2,
		Cursor{Y: 0, X: 3},
		Cursor{Y: 1, X: 1},
		"foo",
	)

	// Blank does nothing
	translateTest(t, 4, 2,
		Cursor{Y: 0, X: 0},
		Cursor{Y: 0, X: 0},
		"",
	)

	// Still nothing
	translateTest(t, 4, 4,
		Cursor{Y: 1, X: 0},
		Cursor{Y: 1, X: 0},
		"foo",
		"",
	)

	// Stays in line
	translateTest(t, 4, 2,
		Cursor{Y: 0, X: 2},
		Cursor{Y: 1, X: 0},
		"foobar",
	)

	// Sent to end
	translateTest(t, 8, 8,
		Cursor{Y: 999, X: 999},
		Cursor{Y: 0, X: 6},
		"foobar",
	)

	// Preserves wrapped state
	translateTest(t, 8, 2,
		Cursor{Y: 0, X: 6},
		Cursor{Y: 2, X: 1, State: cursorWrapNext},
		"foobar",
	)
}

// Double-width characters cannot occupy the last column in a row. Instead, we
// must wrap to a new line instead of just setting cursorWrapNext.
func TestCJKWrap(t *testing.T) {
	term := New()
	term.Resize(4, 2)
	term.Write([]byte(LineFeedMode))
	// Leaves us in the last cell
	term.Write([]byte("foo"))
	// Should cause 你 to be written to the next line instead
	term.Write([]byte("你"))

	lines := term.Screen()
	require.Equal(t, "foo ", lines[0].String())
	require.Equal(t, '你', lines[1][0].Char)
}
