package api

import (
	_ "embed"
	"fmt"
	"io"

	"github.com/cfoust/cy/pkg/bind"
	"github.com/cfoust/cy/pkg/geom"
	"github.com/cfoust/cy/pkg/janet"
	"github.com/cfoust/cy/pkg/mux/screen/tree"
	"github.com/cfoust/cy/pkg/replay"
	"github.com/cfoust/cy/pkg/replay/player"
	"github.com/cfoust/cy/pkg/sessions"
	"github.com/cfoust/cy/pkg/taro"
	"github.com/cfoust/cy/pkg/util"
)

type ReplayModule struct {
	Lifetime             util.Lifetime
	Tree                 *tree.Tree
	TimeBinds, CopyBinds *bind.BindScope
}

func (m *ReplayModule) send(context interface{}, msg taro.Msg) error {
	client, ok := context.(Client)
	if !ok {
		return fmt.Errorf("no client could be inferred")
	}

	node := client.Node()
	if node == nil {
		return fmt.Errorf("client was missing node")
	}

	pane, ok := node.(*tree.Pane)
	if !ok {
		return fmt.Errorf("client node was not pane")
	}

	pane.Screen().Send(msg)
	return nil
}

func (m *ReplayModule) sendAction(context interface{}, action replay.ActionType) error {
	return m.send(context, replay.ActionEvent{
		Type: action,
	})
}

func (m *ReplayModule) sendArg(context interface{}, action replay.ActionType, arg string) error {
	if len(arg) != 1 {
		return nil
	}
	return m.send(context, replay.ActionEvent{
		Type: action,
		Arg:  arg,
	})
}

func (m *ReplayModule) Quit(context interface{}) error {
	return m.sendAction(context, replay.ActionQuit)
}

func (m *ReplayModule) ScrollUp(context interface{}) error {
	return m.sendAction(context, replay.ActionScrollUp)
}

func (m *ReplayModule) ScrollDown(context interface{}) error {
	return m.sendAction(context, replay.ActionScrollDown)
}

func (m *ReplayModule) HalfPageUp(context interface{}) error {
	return m.sendAction(context, replay.ActionScrollUpHalf)
}

func (m *ReplayModule) HalfPageDown(context interface{}) error {
	return m.sendAction(context, replay.ActionScrollDownHalf)
}

func (m *ReplayModule) SearchForward(context interface{}) error {
	return m.sendAction(context, replay.ActionSearchForward)
}

func (m *ReplayModule) TimePlay(context interface{}) error {
	return m.sendAction(context, replay.ActionTimePlay)
}

func (m *ReplayModule) SearchAgain(context interface{}) error {
	return m.sendAction(context, replay.ActionSearchAgain)
}

func (m *ReplayModule) SearchReverse(context interface{}) error {
	return m.sendAction(context, replay.ActionSearchReverse)
}

func (m *ReplayModule) SearchBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionSearchBackward)
}

func (m *ReplayModule) TimeStepBack(context interface{}) error {
	return m.sendAction(context, replay.ActionTimeStepBack)
}

func (m *ReplayModule) TimeStepForward(context interface{}) error {
	return m.sendAction(context, replay.ActionTimeStepForward)
}

func (m *ReplayModule) Beginning(context interface{}) error {
	return m.sendAction(context, replay.ActionBeginning)
}

func (m *ReplayModule) End(context interface{}) error {
	return m.sendAction(context, replay.ActionEnd)
}

func (m *ReplayModule) SwapScreen(context interface{}) error {
	return m.sendAction(context, replay.ActionSwapScreen)
}

func (m *ReplayModule) CursorDown(context interface{}) error {
	return m.sendAction(context, replay.ActionCursorDown)
}

func (m *ReplayModule) CursorLeft(context interface{}) error {
	return m.sendAction(context, replay.ActionCursorLeft)
}

func (m *ReplayModule) CursorRight(context interface{}) error {
	return m.sendAction(context, replay.ActionCursorRight)
}

func (m *ReplayModule) CursorUp(context interface{}) error {
	return m.sendAction(context, replay.ActionCursorUp)
}

func (m *ReplayModule) TimePlaybackRate(context interface{}, rate int) error {
	return m.send(context, replay.PlaybackRateEvent{
		Rate: rate,
	})
}

func (m *ReplayModule) Copy(context interface{}) error {
	return m.sendAction(context, replay.ActionCopy)
}

func (m *ReplayModule) Select(context interface{}) error {
	return m.sendAction(context, replay.ActionSelect)
}

func (m *ReplayModule) JumpAgain(context interface{}) error {
	return m.sendAction(context, replay.ActionJumpAgain)
}

func (m *ReplayModule) JumpReverse(context interface{}) error {
	return m.sendAction(context, replay.ActionJumpReverse)
}

func (m *ReplayModule) JumpBackward(context interface{}, char string) error {
	return m.sendArg(context, replay.ActionJumpBackward, char)
}

func (m *ReplayModule) JumpForward(context interface{}, char string) error {
	return m.sendArg(context, replay.ActionJumpForward, char)
}

func (m *ReplayModule) JumpToForward(context interface{}, char string) error {
	return m.sendArg(context, replay.ActionJumpToForward, char)
}

func (m *ReplayModule) JumpToBackward(context interface{}, char string) error {
	return m.sendArg(context, replay.ActionJumpToBackward, char)
}

func (m *ReplayModule) CommandForward(context interface{}) error {
	return m.sendAction(context, replay.ActionCommandForward)
}

func (m *ReplayModule) CommandBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionCommandBackward)
}

func (m *ReplayModule) CommandSelectForward(context interface{}) error {
	return m.sendAction(context, replay.ActionCommandSelectForward)
}

func (m *ReplayModule) CommandSelectBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionCommandSelectBackward)
}

func (m *ReplayModule) StartOfLine(context interface{}) error {
	return m.sendAction(context, replay.ActionStartOfLine)
}

func (m *ReplayModule) FirstNonBlank(context interface{}) error {
	return m.sendAction(context, replay.ActionFirstNonBlank)
}

func (m *ReplayModule) EndOfLine(context interface{}) error {
	return m.sendAction(context, replay.ActionEndOfLine)
}

func (m *ReplayModule) LastNonBlank(context interface{}) error {
	return m.sendAction(context, replay.ActionLastNonBlank)
}

func (m *ReplayModule) LastNonBlankScreen(context interface{}) error {
	return m.sendAction(context, replay.ActionLastNonBlankScreen)
}

func (m *ReplayModule) FirstNonBlankScreen(context interface{}) error {
	return m.sendAction(context, replay.ActionFirstNonBlankScreen)
}

func (m *ReplayModule) StartOfScreenLine(context interface{}) error {
	return m.sendAction(context, replay.ActionStartOfScreenLine)
}

func (m *ReplayModule) MiddleOfScreenLine(context interface{}) error {
	return m.sendAction(context, replay.ActionMiddleOfScreenLine)
}

func (m *ReplayModule) MiddleOfLine(context interface{}) error {
	return m.sendAction(context, replay.ActionMiddleOfLine)
}

func (m *ReplayModule) EndOfScreenLine(context interface{}) error {
	return m.sendAction(context, replay.ActionEndOfScreenLine)
}

func (m *ReplayModule) WordForward(context interface{}) error {
	return m.sendAction(context, replay.ActionWordForward)
}

func (m *ReplayModule) WordBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionWordBackward)
}

func (m *ReplayModule) WordEndForward(context interface{}) error {
	return m.sendAction(context, replay.ActionWordEndForward)
}

func (m *ReplayModule) WordEndBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionWordEndBackward)
}

func (m *ReplayModule) BigWordForward(context interface{}) error {
	return m.sendAction(context, replay.ActionBigWordForward)
}

func (m *ReplayModule) BigWordBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionBigWordBackward)
}

func (m *ReplayModule) BigWordEndForward(context interface{}) error {
	return m.sendAction(context, replay.ActionBigWordEndForward)
}

func (m *ReplayModule) BigWordEndBackward(context interface{}) error {
	return m.sendAction(context, replay.ActionBigWordEndBackward)
}

func (m *ReplayModule) OpenFile(
	groupId *janet.Value,
	path string,
) (tree.NodeID, error) {
	defer groupId.Free()

	group, err := resolveGroup(m.Tree, groupId)
	if err != nil {
		return 0, err
	}

	reader, err := sessions.Open(path)
	if err != nil {
		return 0, err
	}

	events := make([]sessions.Event, 0)
	for {
		event, err := reader.Read()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return 0, err
		}
		events = append(events, event)
	}

	// TODO(cfoust): 03/04/24 open progress
	ctx := m.Lifetime.Ctx()
	replay := replay.New(
		ctx,
		player.FromEvents(events),
		m.TimeBinds,
		m.CopyBinds,
		replay.WithNoQuit,
	)

	pane := group.NewPane(ctx, replay)
	return pane.Id(), nil
}

type ReplayParams struct {
	Main     bool
	Copy     bool
	Location *geom.Vec2
}

func (m *ReplayModule) Open(
	id *janet.Value,
	named *janet.Named[ReplayParams],
) error {
	defer id.Free()

	pane, err := resolvePane(m.Tree, id)
	if err != nil {
		return err
	}

	var options []replay.Option

	params := named.Values()
	if params.Main {
		options = append(options, replay.WithFlow)
	}

	if params.Copy {
		options = append(options, replay.WithCopyMode)
	}

	if params.Location != nil {
		options = append(options, replay.WithLocation(
			*params.Location,
		))
	}

	r, ok := pane.Screen().(*replay.Replayable)
	if !ok {
		return fmt.Errorf("node not replayable")
	}

	r.EnterReplay(options...)
	return nil
}
