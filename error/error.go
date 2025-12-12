package err

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	Width int
	Height int
	ErrMsg string
}

type ErrSignal struct {
    Msg string
}


func InitError()*model{
	return &model{}
}

func(m model)Init()tea.Cmd{
	 return nil 
}
func(m model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
	case tea.WindowSizeMsg:
        m.Width = msg.Width
		m.Height = msg.Height
        return m, nil
    case ErrSignal:
        m.ErrMsg = msg.Msg
        return m, nil
	}
	
	return m,nil
}
// func (m model) View() string {
// 	bgStyle := lipgloss.NewStyle().Width(m.Width).Height(m.Height).Background(lipgloss.Color("#18181b"))
// 	s := bgStyle.AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).Foreground(lipgloss.Color("#e5e5e5")).Render(fmt.Sprintf("⚠️  oops! an error occured: %s ⚠️",m.ErrMsg))
// 	s1 := bgStyle.AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Bottom).Foreground(lipgloss.Color("#e5e5e5")).Render("press esc to go back")

// 	var b strings.Builder
// 	b.WriteString(s)
// 	b.WriteString(s1)
// 	return b.String()
// }

func (m model) View() string {
    bg := lipgloss.NewStyle().
        Width(m.Width).
        Height(m.Height).
        Background(lipgloss.Color("#18181b"))

    // Main error message
    errText := fmt.Sprintf("⚠️  Oops! an error occurred: %s ⚠️", m.ErrMsg)

    // Footer text
    footerText := "Press esc to go back"

    // Height of both
    errHeight := 1                   // single line
    footerHeight := 1                // single line
    totalContentHeight := errHeight + footerHeight

    // Vertical spacing
    remaining := m.Height - totalContentHeight
    if remaining < 0 {
        remaining = 0
    }

    topPad := remaining / 2
    bottomPad := remaining - topPad

    var b strings.Builder

    // Top empty lines
    b.WriteString(strings.Repeat("\n", topPad))

    // Center ERROR horizontally
    b.WriteString(
        lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, errText),
    )
    b.WriteString("\n")

    // Middle space before footer
    b.WriteString(strings.Repeat("\n", bottomPad))

    // Center FOOTER horizontally
    b.WriteString(
        lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, footerText),
    )

    return bg.Render(b.String())
}
