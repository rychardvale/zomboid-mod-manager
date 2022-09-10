package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
	white    = lipgloss.Color("#FAFAFA")
)

var (
	inputStyle      = lipgloss.NewStyle().Foreground(hotPink)
	cursorStyle     = lipgloss.NewStyle().Bold(true).Foreground(hotPink)
	unselectedStyle = lipgloss.NewStyle().Foreground(darkGray)
	selectedStyle   = lipgloss.NewStyle().Bold(true).Foreground(white)
	ahelpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

const (
	Lista int = iota
	Adicionar
	Remover
	Sair
)

const (
	ModName int = iota
)

// struct that represents a configline in ini file
// values is generated by splitting ";" after "="
type configline struct {
	key    string
	values []string
}

// state of the application
type model struct {
	Choice       int
	Mods         []string
	Cursor       int
	NewMod       []textinput.Model
	Focused      int
	ListStartIdx int // which item is at the top of the list, behaves like a scroll
	ListRange    int // qnty of items in the mod list, defaults to 10

}

func (m *model) SwapMods(from int, to int) {
	temp := m.Mods[to]
	m.Mods[to] = m.Mods[from]
	m.Mods[from] = temp
}

func (m *model) ResetFields() {
	for _, input := range m.NewMod {
		input.Reset()
	}
}

func (m *model) NextInput() {
	m.Focused = (m.Focused + 1) % len(m.NewMod)
}

func (m *model) PrevInput() {
	m.Focused--

	if m.Focused < 0 {
		m.Focused = len(m.NewMod) - 1
	}
}

// checks if the rendered list top item needs to be updated due to scroll function
func (m *model) updateListRange(opts string) {
	switch opts {
	case "down":
		lastitem := len(m.Mods) - 1
		bottomThreshold := m.ListRange - 3
		cantSeeLastItem := m.ListStartIdx+m.ListRange < lastitem
		if (m.Cursor-m.ListStartIdx) > bottomThreshold && cantSeeLastItem {
			m.ListStartIdx++
		}

	case "up":
		if (m.Cursor-m.ListStartIdx) < 3 && m.ListStartIdx > 0 {
			m.ListStartIdx--
		}
	}
}

func (m *model) SubmitMod() {
	m.Focused = ModName
	m.Mods = append(m.Mods, m.NewMod[ModName].Value())
	m.NewMod[ModName].Reset()
}

func initialModel() model {
	configs := loadConfigFile()
	inputs := make([]textinput.Model, 1)
	inputs[ModName] = textinput.New()
	inputs[ModName].Focus()
	inputs[ModName].Placeholder = "Brita_2"
	inputs[ModName].CharLimit = 255

	return model{
		Choice:       Lista,
		Mods:         configs["Mods"],
		Cursor:       0,
		Focused:      0,
		NewMod:       inputs,
		ListStartIdx: 0,
		ListRange:    15,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// update handler for adicionar mod view
func updateAdicionar(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd = make([]tea.Cmd, len(m.NewMod))
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		// when enter is pressed,
		// if its the last input, it is added to the mod list
		// else it is focused on the next input
		case tea.KeyEnter:
			if m.Focused == len(m.NewMod)-1 {
				m.Choice = Lista
				m.Focused = ModName
				m.SubmitMod()
				return m, nil
			} else {
				m.NextInput()
			}

		// exits the view,  goes back to the Lista view
		case tea.KeyCtrlC, tea.KeyEsc:
			m.Choice = Lista
			m.ResetFields()
			return m, nil
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.PrevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.NextInput()
		}
		for i := range m.NewMod {
			m.NewMod[i].Blur()
		}
		m.NewMod[m.Focused].Focus()
	}

	for i := range m.NewMod {
		m.NewMod[i], cmds[i] = m.NewMod[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

// update handler for lista view
func updateLista(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	// fmt.Println(m.Mods)
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "up", "k":
			m.Cursor--
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			m.updateListRange("up")

		case "down", "j":
			m.Cursor++
			lastitem := len(m.Mods) - 1
			if m.Cursor >= lastitem {
				m.Cursor = lastitem
			}
			m.updateListRange("down")

		case "alt+k", "alt+up":
			// loops through the list
			if m.Cursor != 0 {
				m.SwapMods(m.Cursor-1, m.Cursor)
				m.Cursor--
			}
			m.updateListRange("up")


		case "a":
			m.Choice = Adicionar

		case "alt+j", "alt+down":
			if m.Cursor != len(m.Mods)-1 {
				m.SwapMods(m.Cursor+1, m.Cursor)
				m.Cursor++
			}
			m.updateListRange("down")

		case "ctrl+c", "q":
			m.Choice = Sair
			return m, tea.Quit
		}

	}
	return m, nil
}

// main Update function for bubbletea
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Choice = Sair
			return m, tea.Quit
		}
	}

	switch m.Choice {
	case Lista:
		return updateLista(msg, m)
	case Adicionar:
		return updateAdicionar(msg, m)
	default:
		return m, nil
	}
}

// responsible for choosing the right view based on model Choice
func (m model) View() string {
	var s string
	switch m.Choice {

	case Sair:
		s = QuitView(m)
	case Lista:
		s = ListView(m)
	case Adicionar:
		s = AddModView(m)

	}
	return s
}

// loads the config file, TODO: accept path from user to file
func loadConfigFile() map[string][]string {
	data, err := os.ReadFile("./servertest.ini")
	if err != nil {
		log.Fatal(err)
	}
	parsed := strings.Split(string(data), "\n")

	// map containing all file configs,
	// key is the word preceding "=", e.g. "Mods=..." will be "Mods"
	configs := make(map[string][]string)
	for _, s := range parsed {
		if strings.Contains(s, "=") {
			line := ParseConfigLine(&s)
			configs[line.key] = line.values
		}
	}
	return configs
}

func ParseConfigLine(line *string) configline {
	splitted := strings.Split(*line, "=")
	parsedValues := strings.Split(splitted[1], ";")
	return configline{
		key:    splitted[0],
		values: parsedValues,
	}
}

func ListView(m model) string {
	// header
	s := fmt.Sprintf("\n\nInstalled mods: %d\n\n", len(m.Mods))

	for i := m.ListStartIdx; i <= m.ListStartIdx+m.ListRange; i++ {
		if i >= len(m.Mods) {
			break
		}

		cursor := " " // no cursor
		mod := m.Mods[i]
		if m.Cursor == i {
			cursor = cursorStyle.Render(">") // cursor!
			mod = selectedStyle.Render(mod)
		} else {
			mod = unselectedStyle.Render(mod)
		}

		// row
		s += fmt.Sprintf("%s  %s\n", cursor, mod)
	}

	// footer
	s += helpStyle.Render("\n[q] sair.\t[a] adicionar mod.\t[d] deletar mod")

	return s
}

// exiting view
func QuitView(m model) string {
	return fmt.Sprintf(
		"%s",
		cursorStyle.Width(255).Render("So long sucker"),
	)
}

// view for addmod choice
func AddModView(m model) string {
	s := fmt.Sprintf(
		"%s: %s",
		inputStyle.Render("Name"),
		m.NewMod[ModName].View(),
	)

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error brah")
		os.Exit(1)
	}
}