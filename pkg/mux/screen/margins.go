package screen

import (
	"context"

	"github.com/cfoust/cy/pkg/frames"
	"github.com/cfoust/cy/pkg/geom"
	"github.com/cfoust/cy/pkg/geom/tty"
	"github.com/cfoust/cy/pkg/mux"
	"github.com/cfoust/cy/pkg/taro"

	"github.com/sasha-s/go-deadlock"
)

// Margins puts empty space around a Screen and centers it. In size mode,
// Margins attempts to keep the Screen at a fixed size (one or both dimensions
// fixed). In margin mode, the Screen is surrounded by fixed-size margins that
// do not change with the screen size.
type Margins struct {
	deadlock.RWMutex
	*mux.UpdatePublisher

	screen Screen

	// Whether Margins is in margin mode or size mode
	isMargins bool
	margins   Size
	size      Size

	outer Size
	inner geom.Rect
}

var _ Screen = (*Margins)(nil)

func fitMargin(outer, margin int) int {
	if margin == 0 {
		return outer
	}

	return geom.Max(outer-(margin*2), 1)
}

func getSize(outer, desired int) int {
	if desired == 0 {
		return outer
	}

	if desired > outer {
		return outer
	}

	return desired
}

func (l *Margins) Kill() {
	l.RLock()
	screen := l.screen
	l.RUnlock()

	screen.Kill()
}

func (l *Margins) State() *tty.State {
	l.RLock()
	inner := l.inner
	outer := l.outer
	l.RUnlock()

	innerState := l.screen.State()
	state := tty.New(outer)

	tty.Copy(inner.Position, state, innerState)

	size := state.Image.Size()
	for row := 0; row < size.R; row++ {
		for col := 0; col < size.C; col++ {
			if inner.Contains(geom.Vec2{
				R: row,
				C: col,
			}) {
				continue
			}
			state.Image[row][col].Transparent = true
		}
	}

	return state
}

func (l *Margins) Send(msg mux.Msg) {
	l.RLock()
	inner := l.inner
	l.RUnlock()
	l.screen.Send(taro.TranslateMouseMessage(
		msg,
		-inner.Position.C,
		-inner.Position.R,
	))
}

func (l *Margins) getInner(outer Size) geom.Rect {
	// Resolve the desired margins to a real inner window size
	factor := Size{
		R: fitMargin(outer.R, l.margins.R),
		C: fitMargin(outer.C, l.margins.C),
	}

	// Or just ignore it if there's a predetermined size
	if !l.isMargins {
		factor = l.size
	}

	inner := Size{
		R: geom.Max(getSize(outer.R, factor.R), 1),
		C: geom.Max(getSize(outer.C, factor.C), 1),
	}

	return geom.Rect{
		Position: outer.Center(inner),
		Size:     inner,
	}
}

func (l *Margins) SetSize(size Size) error {
	l.Lock()
	l.isMargins = false
	l.size = Size{
		R: geom.Max(0, size.R),
		C: geom.Max(0, size.C),
	}
	l.Unlock()
	return l.recalculate()
}

func (l *Margins) Size() Size {
	l.RLock()
	defer l.RUnlock()
	return l.size
}

func (l *Margins) poll(ctx context.Context) {
	updates := l.screen.Subscribe(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-updates.Recv():
			l.Publish(event)
		}
	}
}

func (l *Margins) recalculate() error {
	l.RLock()
	outer := l.outer
	l.RUnlock()

	inner := l.getInner(outer)

	l.Lock()
	l.inner = inner
	l.Unlock()

	err := l.screen.Resize(inner.Size)
	if err != nil {
		return err
	}

	return nil
}

func (l *Margins) Resize(size Size) error {
	l.Lock()
	l.outer = size
	l.Unlock()
	return l.recalculate()
}

func NewMargins(ctx context.Context, screen Screen) *Margins {
	margins := &Margins{
		UpdatePublisher: mux.NewPublisher(),
		size: Size{
			C: 80,
		},
		screen: screen,
	}

	go margins.poll(ctx)

	return margins
}

func AddMargins(ctx context.Context, screen Screen) mux.Screen {
	innerLayers := NewLayers()
	innerLayers.NewLayer(
		ctx,
		screen,
		PositionTop,
		WithOpaque,
		WithInteractive,
	)
	margins := NewMargins(ctx, innerLayers)

	outerLayers := NewLayers()
	frame := frames.NewFramer(ctx, frames.RandomFrame())
	outerLayers.NewLayer(
		ctx,
		frame,
		PositionTop,
	)

	outerLayers.NewLayer(
		ctx,
		margins,
		PositionTop,
		WithInteractive,
		WithOpaque,
	)

	return outerLayers
}
