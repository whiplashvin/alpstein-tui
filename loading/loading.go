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
}

func InitLoading(width,height int)*Model{
	s := spinner.New()
	s.Spinner = spinner.Meter

	return &Model{
		ScreenName: "Loading",
		Width: width,
		Height: height,
		Spinner: s,
	}
}
func (m Model)Init()tea.Cmd{
	return m.Spinner.Tick
}

func (m Model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
	case spinner.TickMsg:
        var cmd tea.Cmd
        m.Spinner, cmd = m.Spinner.Update(msg)
		return m,cmd
	case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c":
			return m,tea.Quit
		}
	}
	return m,nil
}

func (m Model)View()string{
   spin := m.Spinner.Style.
		Foreground(lipgloss.Color("#ad46ff")).                  
        Align(lipgloss.Center).
        Render(m.Spinner.View()+"  Alpstein loading")

    return lipgloss.Place(
        m.Width,
        m.Height,
        lipgloss.Center, 
        lipgloss.Center,  
        spin,
    )
}

