package fuzzy

import (
	"context"

	"github.com/cfoust/cy/pkg/geom"
	"github.com/cfoust/cy/pkg/geom/tty"
	"github.com/cfoust/cy/pkg/mux"
	"github.com/cfoust/cy/pkg/mux/screen"
	"github.com/cfoust/cy/pkg/util"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/xo/terminfo"
)

type Fuzzy struct {
	util.Lifetime
	*screen.Tea
	result   chan interface{}
	location geom.Vec2
	size     geom.Vec2
}

var _ mux.Screen = (*Fuzzy)(nil)

func (f *Fuzzy) Resize(size mux.Size) error {
	f.size = size
	return nil
}

func (f *Fuzzy) Result() <-chan interface{} {
	return f.result
}

func (f *Fuzzy) State() *tty.State {
	termState := f.Terminal.State()
	state := tty.New(f.size)
	tty.Copy(f.location, state, termState)
	state.CursorVisible = false
	return state
}

func NewFuzzy(
	ctx context.Context,
	profile termenv.Profile,
	info *terminfo.Terminfo,
	options []Option,
	location geom.Vec2,
) *Fuzzy {
	lifetime := util.NewLifetime(ctx)

	result := make(chan interface{})

	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 20
	ti.Prompt = ""

	screen := screen.NewTea(
		ctx,
		model{
			lifetime: &lifetime,
			options:  options,
			result:   result,
			selected: 0,
			renderer: lipgloss.NewRenderer(
				nil,
				termenv.WithProfile(profile),
			),
			textInput: ti,
		},
		profile,
		info,
		geom.DEFAULT_SIZE,
	)

	return &Fuzzy{
		Lifetime: lifetime,
		Tea:      screen,
		result:   result,
		location: location,
	}
}

type model struct {
	lifetime  *util.Lifetime
	renderer  *lipgloss.Renderer
	textInput textinput.Model
	result    chan interface{}

	options  []Option
	filtered []Option
	selected int
	pattern  string
}

var _ tea.Model = (*model)(nil)

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

type matchResult struct {
	Filtered []Option
}

func queryOptions(options []Option, pattern string) tea.Cmd {
	return func() tea.Msg {
		return matchResult{
			Filtered: Filter(options, pattern),
		}
	}
}

func (m model) getOptions() []Option {
	if len(m.pattern) > 0 {
		return m.filtered
	}
	return m.options
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case matchResult:
		m.filtered = msg.Filtered
		m.selected = geom.Max(geom.Min(m.selected, len(m.getOptions())-1), 0)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			m.result <- nil
			m.lifetime.Cancel()
			return m, nil
		case tea.KeyUp, tea.KeyCtrlK:
			m.selected = geom.Max(m.selected-1, 0)
			return m, nil
		case tea.KeyDown, tea.KeyCtrlJ:
			m.selected = geom.Min(m.selected+1, len(m.getOptions())-1)
			return m, nil
		case tea.KeyEnter:
			if m.selected < len(m.getOptions()) {
				option := m.getOptions()[m.selected]
				m.result <- option.Result
			} else {
				m.result <- nil
			}
			m.lifetime.Cancel()
			return m, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	value := m.textInput.Value()
	if m.pattern != value {
		m.pattern = value
		cmds = append(cmds, queryOptions(m.options, value))
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	basic := m.renderer.NewStyle().
		Background(lipgloss.Color("#20111B")).
		Foreground(lipgloss.Color("#D5CCBA")).
		Width(20)

	inactive := basic.Copy().Background(lipgloss.Color("#968C83"))
	active := basic.Copy().
		Background(lipgloss.Color("#EAA549")).
		Foreground(lipgloss.Color("#20111B"))

	var options []string
	for i, match := range m.getOptions() {
		var rendered string
		if m.selected == i {
			rendered = active.Render(match.Text)
		} else {
			rendered = inactive.Render(match.Text)
		}
		options = append(options, rendered)
	}

	m.textInput.Cursor.Style = m.renderer.NewStyle().
		Background(lipgloss.Color("#EAA549"))

	return lipgloss.JoinVertical(lipgloss.Left,
		basic.Render(m.textInput.View()),
		lipgloss.JoinVertical(lipgloss.Left, options...),
	)
}
