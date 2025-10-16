package golater

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C3C3C")).
			Padding(0, 1).
			Render("│")

	previewBoxStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5B5B5B")).
			MarginLeft(1).
			Width(45)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	fieldNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#B984BE")).
			Padding(0, 1)

	inputContainerStyle = lipgloss.NewStyle().
				MarginLeft(2).
				MarginBottom(1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type GolaterConfig struct {
	Path     string     `json:"path"`
	Template []Template `json:"templates"`
}

type Template struct {
	Name          string   `json:"name"`
	Desc          string   `json:"description"`
	Lang          string   `json:"lang"`
	EndlineFormat string   `json:"endline_format"`
	Repeated      Repeated `json:"repeated"`
	Root          []File   `json:"root"`
}

type Repeated struct {
	Folder string `json:"folder"`
	Files  []File `json:"files"`
}

type File struct {
	Filename string   `json:"filename"`
	Ext      string   `json:"ext"`
	Data     []string `json:"data"`
}

func (t Template) Title() string       { return t.Name }
func (t Template) Description() string { return t.Desc }
func (t Template) FilterValue() string { return t.Name }

type ViewMode int

const (
	MainPage ViewMode = iota
	SpawnPage
)

var ViewModes = struct {
	MainPage  ViewMode
	SpawnPage ViewMode
}{
	MainPage:  MainPage,
	SpawnPage: SpawnPage,
}

func readTemplatesCmd() tea.Cmd {
	return func() tea.Msg {
		ch := make(chan string)
		var cfg GolaterConfig
		var err error

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg, err = ReadTemplates(ch)
			if err != nil {
				ch <- fmt.Sprintf("Error: %v", err)
			}
			close(ch)
		}()

		for msg := range ch {
			fmt.Println(msg)
		}

		// var b strings.Builder
		// for _, v := range cfg.Template {
		// 	b.Write([]byte(fmt.Sprintf("")))
		// }

		return cfg
	}
}

func spawnTemplateCmd(t Template, iters int) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan string)
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			err := SpawnTemplate(t, iters, ch)
			if err != nil {
				ch <- fmt.Sprintf("Error: %v", err)
			}
			close(ch)
		}()

		var b strings.Builder
		for msg := range ch {
			b.WriteString(msg)
		}

		wg.Wait()
		return tea.Quit()
	}
}

type listKeyMap struct {
	editItem    key.Binding
	deleteItem  key.Binding
	spawnItem   key.Binding
	showPreview key.Binding
}

func createListKeyMap() *listKeyMap {
	return &listKeyMap{
		spawnItem: key.NewBinding(
			key.WithHelp("enter", "spawn"),
			key.WithKeys("enter"),
		),
		editItem: key.NewBinding(
			key.WithHelp("ctrl+e", "edit"),
			key.WithKeys("ctrl+e"),
		),
		showPreview: key.NewBinding(
			key.WithHelp("h", "show preview"),
			key.WithKeys("h"),
		),
	}
}

type model struct {
	viewMode     ViewMode
	keys         *listKeyMap
	core         GolaterConfig
	templateList list.Model
	cursor       int
	preview      Preview
	spawnIters   int
	iterInput    IterInput
}

type IterInput struct {
	fieldName string
	input     textinput.Model
}

type Preview struct {
	show_preview bool
	viewport     viewport.Model
}

func SetupDefaultModel() model {
	msg := readTemplatesCmd()()
	cfg, ok := msg.(GolaterConfig)
	if !ok {
		cfg = GolaterConfig{}
	}

	keyList := createListKeyMap()

	tlist := make([]list.Item, len(cfg.Template))
	for i := range cfg.Template {
		tlist[i] = cfg.Template[i]
	}

	ti := textinput.New()
	ti.Placeholder = "enter number"
	ti.CharLimit = 3
	ti.Width = 20

	delegate := list.NewDefaultDelegate()
	templateList := list.New(tlist, delegate, 0, 0)
	templateList.Title = "Existing templates"
	templateList.Styles.Title = titleStyle
	templateList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyList.spawnItem,
			keyList.editItem,
			keyList.deleteItem,
		}
	}
	templateList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyList.showPreview,
		}
	}

	preview := viewport.New(100, 20)

	model := model{
		viewMode:     ViewModes.MainPage,
		keys:         keyList,
		core:         cfg,
		templateList: templateList,
		cursor:       0,
		spawnIters:   0,
		iterInput: IterInput{
			fieldName: "Enter iterations",
			input:     ti,
		},
		preview: Preview{
			false,
			preview,
		},
	}
	return model
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// h, v := appStyle.GetFrameSize()

		m.templateList.SetSize(msg.Width/2-2, msg.Height-3)
		m.preview.viewport.Width = msg.Width/2 - 2
		m.preview.viewport.Height = msg.Height - 2

	case tea.KeyMsg:
		switch msg.String() {
		case "down", "s", "j":
			if m.cursor+1 < len(m.core.Template) {
				m.cursor++
			}
		case "up", "w", "k":
			if m.cursor-1 >= 0 {
				m.cursor--
			}
		case "h":
			if m.viewMode == ViewModes.MainPage {
				m.preview.show_preview = !m.preview.show_preview
			}
		case "enter":
			switch m.viewMode {
			case ViewModes.MainPage:
				m.viewMode = ViewModes.SpawnPage
				m.iterInput.input.Focus()
			case ViewModes.SpawnPage:
				if m.spawnIters > 0 {
					return m, spawnTemplateCmd(m.core.Template[m.cursor], m.spawnIters)
				}
			}
		}

		lipgloss.JoinHorizontal(lipgloss.Right, GeneratePreviewForTemplate(m.core.Template[m.cursor]))

		switch m.viewMode {
		case ViewModes.MainPage:
			var cmd tea.Cmd
			m.templateList, cmd = m.templateList.Update(msg)
			cmds = append(cmds, cmd)
		case ViewModes.SpawnPage:
			var cmd tea.Cmd
			m.iterInput.input, cmd = m.iterInput.input.Update(msg)

			inputValue := m.iterInput.input.Value()
			digitsOnly := strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				if r == 'q' {
					cmd = tea.Quit
				}
				return -1
			}, inputValue)

			if inputValue != digitsOnly {
				m.iterInput.input.SetValue(digitsOnly)
			}

			if n, err := strconv.Atoi(digitsOnly); err == nil {
				m.spawnIters = n
			}
			if digitsOnly == "" {
				m.spawnIters = 1
			}

			cmds = append(cmds, cmd)
		}

	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var renderString string = ""
	switch m.viewMode {
	case ViewModes.MainPage:
		description := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B7B7B")).
			MarginTop(1).
			MarginBottom(1).
			MarginLeft(2).
			Render("enter to spawn | templates loaded from " + m.core.Path)

		left := lipgloss.NewStyle().
			Width(m.templateList.Width()).
			Height(m.templateList.Height() + 3).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Top,
					description,
					m.templateList.View(),
				),
			)

		right := ""
		if m.preview.show_preview {
			if len(m.core.Template) > 0 {
				right = GeneratePreviewForTemplate(m.core.Template[m.cursor])
				m.preview.viewport.SetContent(right)
				right = m.preview.viewport.View()
			} else {
				right = lipgloss.NewStyle().Italic(true).Render("no templates loaded")
			}
		}

		renderString = lipgloss.JoinHorizontal(
			lipgloss.Top,
			left,
			right,
		)
	case ViewModes.SpawnPage:

		field := lipgloss.NewStyle().MarginBottom(1).Render(m.iterInput.input.View())
		title := fieldNameStyle.Render(m.iterInput.fieldName)
		renderString = fmt.Sprintf("%s\n  %s\n", title, field)

		renderString = inputContainerStyle.Render(renderString)
	}

	return renderString
}

func (m model) Init() tea.Cmd {
	return nil
}

func Golater() *tea.Program {
	l, err := NewLogger(zapcore.InfoLevel.String(), "golater.log")
	if err != nil {
		log.Fatalf("failed creating logger | err: %v", err)
	}
	ReplaceGlobals(l)

	model := SetupDefaultModel()
	tp := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err != nil {
		zap.L().Error("failed starting app", zap.Error(err))
	}
	return tp
}
