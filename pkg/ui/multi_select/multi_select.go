package multiselect

import (
	"fmt"

	"github.com/AntoninoAdornetto/issue-summoner/pkg/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type Selection struct {
	Options map[string]bool
}

type Step struct {
	StepName string
	Options  []Item
	Headers  string
	Field    string
}

type Item struct {
	ID, Title, Desc string
}

func (s *Selection) OnSelect(option string, value bool) {
	s.Options[option] = value
}

type model struct {
	cursor   int
	options  []Item
	selected map[int]struct{}
	choices  *Selection
	header   string
	exit     *bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func InitialModelMultiSelect(
	options []Item,
	selection *Selection,
	header string,
	program *bool,
) model {
	return model{
		options:  options,
		selected: make(map[int]struct{}),
		choices:  selection,
		header:   ui.AccentTextStyle.Render(header),
		exit:     program,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			*m.exit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "y":
			for selectedKey := range m.selected {
				m.choices.OnSelect(m.options[selectedKey].ID, true)
				m.cursor = selectedKey
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := m.header + "\n\n"

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ui.SuccessTextStyle.Render(">")
			option.Title = ui.PrimaryTextStyle.Render(option.Title)
			option.Desc = ui.PrimaryTextStyle.Render(option.Desc)
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = ui.SecondaryTextStyle.Render("*")
		}

		title := ui.DimTextStyle.Render(option.Title)
		description := ui.DimTextStyle.Render(option.Desc)

		s += fmt.Sprintf("%s [%s] %s\n%s\n\n", cursor, checked, title, description)
	}

	s += fmt.Sprintf("Press %s to confirm choice.\n", ui.AccentTextStyle.Render("y"))
	return s
}