package replay

import (
	"time"

	"github.com/cfoust/cy/pkg/emu"
	"github.com/cfoust/cy/pkg/geom"
	P "github.com/cfoust/cy/pkg/io/protocol"
	"github.com/cfoust/cy/pkg/taro"

	tea "github.com/charmbracelet/bubbletea"
)

// Move the terminal back in time to the event at `index` and byte offset (if
// the event is an OutputMessage) of `indexByte`.
func (r *Replay) setIndex(index, indexByte int, updateTime bool) {
	numEvents := len(r.events)
	// Allow for negative indices from end of stream
	if index < 0 {
		index = geom.Clamp(numEvents+index, 0, numEvents-1)
	}

	fromIndex := geom.Clamp(r.location.Index, 0, numEvents-1)
	toIndex := geom.Clamp(index, 0, numEvents-1)
	fromByte := r.location.Offset
	toByte := indexByte

	// Going back in time; must start over
	if toIndex < fromIndex || (toIndex == fromIndex && toByte < fromByte) {
		r.terminal = emu.New()
		fromIndex = 0
		fromByte = -1
	}

	for i := fromIndex; i <= toIndex; i++ {
		event := r.events[i]
		switch e := event.Message.(type) {
		case P.OutputMessage:
			data := e.Data

			if toIndex == i {
				if toByte < 0 {
					toByte += len(data)
				}
				toByte = geom.Clamp(toByte, 0, len(data)-1)
			}

			if fromIndex == toIndex {
				data = data[fromByte+1 : toByte+1]
			} else if fromIndex == i {
				data = data[fromByte+1:]
			} else if toIndex == i {
				data = data[:toByte+1]
			}

			r.terminal.Write(data)
		case P.SizeMessage:
			r.terminal.Resize(
				e.Columns,
				e.Rows,
			)
		}
	}

	r.location.Index = toIndex
	r.location.Offset = toByte
	if updateTime {
		r.currentTime = r.events[toIndex].Stamp
	}

	r.recalculateViewport()
	termCursor := r.getTerminalCursor()
	viewportCursor := r.termToViewport(termCursor)

	r.isSelectionMode = false

	// reset scroll offset whenever we move in time
	r.offset.R = 0
	r.offset.C = 0

	// Center the cursor if it's not in the viewport
	if !r.isInViewport(viewportCursor) {
		r.center(termCursor)
	}

	r.cursor = r.termToViewport(termCursor)
	r.desiredCol = r.cursor.C
}

func (r *Replay) gotoIndex(index, indexByte int) {
	r.isPlaying = false
	r.setIndex(index, indexByte, true)
}

func (r *Replay) gotoMatch(index int) {
	if len(r.matches) == 0 {
		return
	}

	index = geom.Clamp(index, 0, len(r.matches)-1)
	r.matchIndex = index
	match := r.matches[index].Begin
	r.gotoIndex(match.Index, match.Offset)
}

func (r *Replay) gotoMatchDelta(delta int) {
	numMatches := len(r.matches)
	if numMatches == 0 {
		return
	}

	// Python-esque modulo behavior
	index := (((r.matchIndex + delta) % numMatches) + numMatches) % numMatches
	r.gotoMatch(index)
}

func (r *Replay) scheduleUpdate() (taro.Model, tea.Cmd) {
	since := time.Now()
	return r, func() tea.Msg {
		time.Sleep(time.Second / PLAYBACK_FPS)
		return PlaybackEvent{
			Since: since,
		}
	}
}

func (r *Replay) setTimeDelta(delta time.Duration) {
	if len(r.events) == 0 {
		return
	}

	newTime := r.currentTime.Add(delta)
	if newTime.Equal(r.currentTime) {
		return
	}

	beginning := r.events[0].Stamp
	lastIndex := len(r.events) - 1
	end := r.events[lastIndex].Stamp
	if newTime.Before(beginning) || newTime.Equal(beginning) {
		r.gotoIndex(0, -1)
		return
	}

	if newTime.After(end) || newTime.Equal(end) {
		r.gotoIndex(lastIndex, -1)
		return
	}

	// We use setIndex after this because our timestamp can be anywhere
	// within the valid range; gotoIndex sets the time to the timestamp of
	// the event
	if newTime.Before(r.currentTime) {
		indexStamp := r.events[r.location.Index].Stamp
		for i := r.location.Index; i >= 0; i-- {
			if newTime.Before(indexStamp) && newTime.After(r.events[i].Stamp) {
				r.setIndex(i, -1, false)
				break
			}
		}
	} else {
		for i := r.location.Index + 1; i < len(r.events); i++ {
			if newTime.After(r.events[i].Stamp) {
				r.setIndex(i, -1, false)
				break
			}
		}
	}

	r.currentTime = newTime
}
