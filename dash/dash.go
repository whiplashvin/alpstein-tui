package dash

import (
	"fmt"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


type model struct{
	ScreenName string
	Jwt interface{}
	Url string
	Width int
	Height int
	CurrUser string
	Cryptos string
	ErrorModel tea.Model
}

type Cryptos string

func InitDash(jwt interface{},url,currUser string,width,height int)*model{
	return &model{
		ScreenName: "Dash",
		Jwt: jwt,
		Url: url,
		Width: width,
		Height: height,
		CurrUser: currUser,
	}
}

func (m model)Init()tea.Cmd{
	return m.FetchLiveCryptos()
}

func (m model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
	case Cryptos:
		m.Cryptos = string(msg)
	case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c":
			return m,tea.Quit
		case "esc":

		}
	}
	return m,nil
}

func (m model)View()string{
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Width(m.Width).Align(lipgloss.Center).MarginTop(2).MarginBottom(5)
	s := fmt.Sprintf("Hey %s! Welcome to Alpstein\n %s",m.CurrUser,m.Cryptos)
	styledHeading := style.Render(s)
	return styledHeading
}
func(m *model)FetchLiveCryptos()tea.Cmd{
	return func() tea.Msg {
			req,err := http.NewRequest(http.MethodGet,fmt.Sprintf("%s/live-cryptos?limit=1",m.Url),nil)
	if err != nil{

	}
	req.Header.Set("Authorization",fmt.Sprintf("Bearer %s",m.Jwt))
	client := http.Client{
		 Timeout: 10 * time.Second,
	}
	resp,err := client.Do(req)
	if err != nil{
		
	}
	defer resp.Body.Close()

	b,_ := io.ReadAll(resp.Body)
	return Cryptos(string(b))
	}
}