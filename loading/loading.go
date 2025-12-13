package loading

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct{
	CurrModel *tea.Model
	ScreenName string
	Width int
	Height int
	Spinner spinner.Model
	bgColor string
	primaryTextColor string
	secondaryTextColor string
}

type LoadingSignal struct{}

func InitLoading()*Model{
	s := spinner.New()
	s.Spinner = spinner.Meter

	return &Model{
		ScreenName: "Loading",
		Spinner: s,
		bgColor: "#18181b",
		primaryTextColor: "#a3b3ff",
		secondaryTextColor: "#c7d8ff",
	}
}
func (m Model)Init()tea.Cmd{
	return m.Spinner.Tick
}

func (m Model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
	case tea.WindowSizeMsg:
        m.Width = msg.Width
		m.Height = msg.Height
        return m, nil
	case spinner.TickMsg:
        var cmd tea.Cmd
        m.Spinner, cmd = m.Spinner.Update(msg)
		return m,cmd
	case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c":
			return m,tea.Quit
		}
	case LoadingSignal:
		return m,nil
	}
	return m,nil
}

func (m Model)View()string{
   spin := m.Spinner.Style.
		Height(m.Height).
		Width(m.Width).
   		Background(lipgloss.Color(m.bgColor)).
		Foreground(lipgloss.Color(m.primaryTextColor)).                  
        AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
        Render(m.Spinner.View()+"  Alpstein loading")

    return lipgloss.Place(
        m.Width,
        m.Height,
        lipgloss.Center, 
        lipgloss.Center,  
        spin,
    )
}

