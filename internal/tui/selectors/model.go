package selectors

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Item struct {
	FieldTitle string
	FieldDesc  string
}

func (i Item) Title() string       { return i.FieldTitle }
func (i Item) Description() string { return i.FieldDesc }
func (i Item) FilterValue() string { return i.FieldTitle }

type Model struct {
	list list.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return docStyle.Render(m.list.View())
}

func Init() (tea.Model, error) {
	items := []list.Item{
		Item{FieldTitle: "Raspberry Pi’s", FieldDesc: "I have ’em all over my house"},
		Item{FieldTitle: "Nutella", FieldDesc: "It's good on toast"},
		Item{FieldTitle: "Bitter melon", FieldDesc: "It cools you down"},
		Item{FieldTitle: "Nice socks", FieldDesc: "And by that I mean socks without holes"},
		Item{FieldTitle: "Eight hours of sleep", FieldDesc: "I had this once"},
		Item{FieldTitle: "Cats", FieldDesc: "Usually"},
		Item{FieldTitle: "Plantasia, the album", FieldDesc: "My plants love it too"},
		Item{FieldTitle: "Pour over coffee", FieldDesc: "It takes forever to make though"},
		Item{FieldTitle: "VR", FieldDesc: "Virtual reality...what is there to say?"},
		Item{FieldTitle: "Noguchi Lamps", FieldDesc: "Such pleasing organic forms"},
		Item{FieldTitle: "Linux", FieldDesc: "Pretty much the best OS"},
		Item{FieldTitle: "Business school", FieldDesc: "Just kidding"},
		Item{FieldTitle: "Pottery", FieldDesc: "Wet clay is a great feeling"},
		Item{FieldTitle: "Shampoo", FieldDesc: "Nothing like clean hair"},
		Item{FieldTitle: "Table tennis", FieldDesc: "It’s surprisingly exhausting"},
		Item{FieldTitle: "Milk crates", FieldDesc: "Great for packing in your extra stuff"},
		Item{FieldTitle: "Afternoon tea", FieldDesc: "Especially the tea sandwich part"},
		Item{FieldTitle: "Stickers", FieldDesc: "The thicker the vinyl the better"},
		Item{FieldTitle: "20° Weather", FieldDesc: "Celsius, not Fahrenheit"},
		Item{FieldTitle: "Warm light", FieldDesc: "Like around 2700 Kelvin"},
		Item{FieldTitle: "The vernal equinox", FieldDesc: "The autumnal equinox is pretty good too"},
		Item{FieldTitle: "Gaffer’s tape", FieldDesc: "Basically sticky fabric"},
		Item{FieldTitle: "Terrycloth", FieldDesc: "In other words, towel fabric"},
	}

	m := Model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "My Fave Things"

	p := tea.NewProgram(m)

	model, err := p.Run()
	if err != nil {
		return nil, err
	}

	return model, nil
}
