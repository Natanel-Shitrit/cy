package player

import (
	"github.com/cfoust/cy/pkg/emu"
	"github.com/cfoust/cy/pkg/geom"
	P "github.com/cfoust/cy/pkg/io/protocol"
	"github.com/cfoust/cy/pkg/sessions"
	"github.com/cfoust/cy/pkg/sessions/search"

	"github.com/sasha-s/go-deadlock"
)

const (
	CY_HOOK = "cy"
)

type Player struct {
	emu.Terminal
	mu       deadlock.RWMutex
	inUse    bool
	buffer   []sessions.Event
	events   []sessions.Event
	location search.Address
}

var _ sessions.EventHandler = (*Player)(nil)

func (p *Player) Acquire() {
	p.mu.Lock()
	p.inUse = true
	p.mu.Unlock()
}

func (p *Player) resetTerminal() {
	p.Terminal = emu.New()
	p.Terminal.Changes().SetHooks([]string{CY_HOOK})
}

func (p *Player) consume(event sessions.Event) {
	p.events = append(p.events, event)
}

func (p *Player) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.inUse = false
	buffer := p.buffer
	p.buffer = make([]sessions.Event, 0)

	// Bring us up-to-date
	p.Goto(len(p.events)-1, -1)

	for _, event := range buffer {
		p.consume(event)
	}
}

func (p *Player) Events() []sessions.Event {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.events
}

func (p *Player) Location() search.Address {
	return p.location
}

func (p *Player) IsAltMode() bool {
	return emu.IsAltMode(p.Terminal.Mode())
}

func (p *Player) Goto(index, offset int) {
	numEvents := len(p.events)
	if numEvents == 0 {
		return
	}

	// Allow for negative indices from end of stream
	if index < 0 {
		index = geom.Clamp(numEvents+index, 0, numEvents-1)
	}

	fromIndex := geom.Clamp(p.location.Index, 0, numEvents-1)
	toIndex := geom.Clamp(index, 0, numEvents-1)
	fromByte := p.location.Offset
	toByte := offset

	// Going back in time; must start over
	if toIndex < fromIndex || (toIndex == fromIndex && toByte < fromByte) {
		p.resetTerminal()
		fromIndex = 0
		fromByte = -1
	}

	for i := fromIndex; i <= toIndex; i++ {
		event := p.events[i]
		switch e := event.Message.(type) {
		case P.OutputMessage:
			data := e.Data

			if toIndex == i {
				if toByte < 0 {
					toByte += len(data)
				}
				toByte = geom.Clamp(toByte, 0, len(data)-1)
			}

			if len(data) > 0 {
				if fromIndex == toIndex {
					data = data[fromByte+1 : toByte+1]
				} else if fromIndex == i {
					data = data[fromByte+1:]
				} else if toIndex == i {
					data = data[:toByte+1]
				}
			}

			p.Terminal.Parse(data)
		case P.SizeMessage:
			p.Terminal.Resize(e.Vec())
		}
	}

	p.location.Index = toIndex
	p.location.Offset = toByte
	return
}

func (p *Player) Process(event sessions.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	inUse := p.inUse
	if !inUse {
		p.consume(event)
		return nil
	}

	p.buffer = append(p.buffer, event)
	return nil
}

func New() *Player {
	p := &Player{}
	p.resetTerminal()
	return p
}

func FromEvents(events []sessions.Event) *Player {
	player := New()

	for _, event := range events {
		player.Process(event)
	}

	return player
}
