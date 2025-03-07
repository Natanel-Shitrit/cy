package movement

import (
	"testing"

	"github.com/cfoust/cy/pkg/emu"
	"github.com/cfoust/cy/pkg/geom"
	"github.com/cfoust/cy/pkg/geom/tty"

	"github.com/stretchr/testify/require"
)

func TestHighlight(
	t *testing.T,
	m Movement,
	size geom.Size,
	highlights []Highlight,
	lines ...string,
) {
	bg := emu.Color(1)
	for i := range highlights {
		highlights[i].BG = bg
	}

	state := tty.New(size)
	m.View(state, highlights)
	image := state.Image

	for row := range lines {
		for col, char := range lines[row] {
			switch char {
			case '0':
				require.NotEqual(
					t,
					bg,
					image[row][col].BG,
					"cell [%d, %d] should not be highlighted",
					row,
					col,
				)
			case '1':
				require.Equal(
					t,
					bg,
					image[row][col].BG,
					"cell [%d, %d] should be highlighted",
					row,
					col,
				)
			default:
				t.Logf(
					"invalid char %+v, must be 1 or 0",
					char,
				)
				t.FailNow()
			}
		}
	}
}
