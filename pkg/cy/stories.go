package cy

import (
	"context"

	"github.com/cfoust/cy/pkg/geom"
	"github.com/cfoust/cy/pkg/mux"
	S "github.com/cfoust/cy/pkg/mux/screen"
	"github.com/cfoust/cy/pkg/stories"
)

func createStory(ctx context.Context) (cy *Cy, client *Client, screen mux.Screen, err error) {
	cy, err = Start(ctx, Options{
		Shell:      "/bin/bash",
		HideSplash: true,
	})
	if err != nil {
		return
	}

	client, err = cy.NewClient(ctx, ClientOptions{
		Env: map[string]string{
			"TERM": "xterm-256color",
		},
		Size: geom.DEFAULT_SIZE,
	})
	if err != nil {
		return
	}

	screen = S.NewTerminal(
		ctx,
		client,
		geom.DEFAULT_SIZE,
	)

	return
}

var initWithFrame stories.InitFunc = func(ctx context.Context) (mux.Screen, error) {
	_, _, screen, err := createStory(ctx)
	return screen, err
}

var initNoFrame stories.InitFunc = func(ctx context.Context) (mux.Screen, error) {
	_, client, screen, err := createStory(ctx)
	client.execute(`(viewport/set-size [0 0])`)
	return screen, err
}

func init() {
	stories.Register("cy/viewport", initWithFrame, stories.Config{
		Input: []interface{}{
			stories.Wait(stories.Some),
			stories.Type("ctrl+a", "g"),
			stories.Wait(stories.More),
			stories.Type("ctrl+a", "g"),
			stories.Wait(stories.More),
			stories.Type("ctrl+a", "+"),
			stories.Wait(stories.Some),
			stories.Type("ctrl+a", "-"),
			stories.Wait(stories.More),
		},
	})

	stories.Register("cy/replay", initNoFrame, stories.Config{
		Input: []interface{}{
			stories.Wait(stories.Some),
			stories.Type("this is a test", "enter"),
			stories.Wait(stories.More),
			stories.Type("seq 1 10", "enter"),
			stories.Wait(stories.More),
			stories.Type("ctrl+a", "p"),
			stories.Wait(stories.Some),
			stories.Type("left", "left", "left", "left"),
			stories.Wait(stories.ALot),
			stories.Type("space"),
			stories.Wait(stories.ALot),
			stories.Type("?", "test", "enter"),
			stories.Wait(stories.ALot),
			stories.Type("h", "h", "h"),
			stories.Wait(stories.ALot),
		},
	})

	stories.Register("cy/shell", initNoFrame, stories.Config{
		Input: []interface{}{
			stories.Wait(stories.Some),
			stories.Type("ls -lah", "enter"),
			stories.Wait(stories.More),
			stories.Type("this is the first shell"),
			stories.Wait(stories.More),
			stories.Type("ctrl+a", "j"),
			stories.Wait(stories.More),
			stories.Type("this is a new shell"),
			stories.Wait(stories.ALot),
			stories.Type("ctrl+l"),
			stories.Wait(stories.Some),
			stories.Type("ctrl+l"),
			stories.Wait(stories.ALot),
		},
	})
}
