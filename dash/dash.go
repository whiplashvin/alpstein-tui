package dash

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CryptoModel struct {
	Id		          string        `gorm:"primaryKey" json:"id"`
	SourceUrl         string        `json:"sourceurl"`
	Heading           string        `json:"heading"`
	Name   	          string        `json:"name"`
	Symbol            string        `json:"symbol"`
	Synopsis          string        `json:"synopsis"`
	Position          string        `json:"position"`
	Buy               string        `json:"buy"`
	BuyPrice          float64       `json:"buyprice"`
	TakeProfit        float64       `json:"takeprofit"`
	StopLoss          float64       `json:"stoploss"`
	Sell	          string        `json:"sell"`
	SellPrice  	      float64       `json:"sellprice"`
	ShortCoverProfit  float64       `json:"shortcoverprofit,omitempty"`
	ShortCoverLoss    float64       `json:"shortcoverloss,omitempty"`
	WaitOut           string        `json:"waitout"`
	Monitor           string        `json:"monitor"`
	Tag               string        `json:"tag"`
	PriceAtCreation   float64       `json:"priceAtCreation"`
	TriggeredPosition string        `json:"triggeredposition"`
	Status            string        `json:"status"`
	ScrappedAt        int64         `json:"scrappedat"`
	CreatedAt         int64         `json:"createdat"`
	TriggeredAt		  int64			`json:"triggeredat"`
	ClosureAt		  int64			`json:"closureat"`
}

type AllCryptoResponse struct{
	Data []CryptoModel `json:"data"`
}
type CryptoAbout struct{
	Id 	   string `json:"id"`
	Symbol string `json:"symbol"`
	About  string `json:"about"`
}
type SingleCryptoResponse struct{
	Data []json.RawMessage `json:"data"`
}
type model struct{
	ScreenName string
	bgColor string
	primaryTextColor string
	secondaryTextColor string
	Width int
	Height int
	Jwt interface{}
	Url string
	Cursor int
	CurrUser string
	Cryptos []CryptoModel
	CurrCryptoId string
	ErrorModel tea.Model
	CurrCrypto CryptoModel
	debounceID int
}

type DebounceFetch struct {
	id int
}
type Cryptos []CryptoModel
type SetCryptoId string
type SetCurrCrypto CryptoModel

func debounceCmd(id int, delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return DebounceFetch{id: id}
	})
}


func InitDash(jwt string,url,currUser string,width,height int)*model{
	return &model{
		ScreenName: "Dash",
		// bgColor: "#18181b",
		primaryTextColor: "#a3b3ff",
		secondaryTextColor: "#c7d8ff",
		Width: width,
		Height: height,
		Jwt: jwt,
		Url: url,
		Cursor: 0,
		CurrUser: currUser,
	}
}

func (m model)Init()tea.Cmd{
	return m.FetchLiveCryptos() 
}

func (m model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
	case DebounceFetch:
		if msg.id != m.debounceID {
			return m, nil
		}
		return m, func() tea.Msg {
			return SetCryptoId(m.Cryptos[m.Cursor].Id)
		}
	case SetCryptoId:
		m.CurrCryptoId = string(msg)
		cmd := m.fetchCryptoByID()
		return m, cmd
	case SetCurrCrypto:	
		m.CurrCrypto = CryptoModel(msg)
	case Cryptos:
		m.Cryptos = msg
		m.CurrCryptoId = m.Cryptos[0].Id
		m.CurrCrypto = m.Cryptos[0]
	case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c":
			return m,tea.Quit
		case "esc":
			return m,tea.Quit
		case "down":
			if m.Cursor < len(m.Cryptos)-1{
				m.Cursor++
				m.debounceID++
				return m, debounceCmd(m.debounceID, 500 * time.Millisecond)
				// return m, func() tea.Msg {
				// 	return SetCryptoId(m.Cryptos[m.Cursor].Id)
				// }
			}
		case "up":
			if m.Cursor > 0 {
				m.Cursor--
				m.debounceID++
				return m, debounceCmd(m.debounceID, 500 * time.Millisecond)
				// return m, func() tea.Msg {
				// 	return SetCryptoId(m.Cryptos[m.Cursor].Id)
				// }
			}
	}
		
	}
	return m,nil
}

func (m model)View()string{
	bg := lipgloss.NewStyle().Width(m.Width).Height(m.Height)

	left := lipgloss.NewStyle().
	Width((m.Width / 2)-2).
	AlignHorizontal(lipgloss.Left).
	Foreground(lipgloss.Color(m.secondaryTextColor)).MarginLeft(2).MarginTop(1).
	Render("ALPSTEIN")

	right := lipgloss.NewStyle().
	Width(m.Width / 2).
	AlignHorizontal(lipgloss.Right).
	Foreground(lipgloss.Color(m.primaryTextColor)).MarginTop(1).
	Render(fmt.Sprintf("hey %s!", m.CurrUser))

	heading := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)

	headingHeight := lipgloss.Height(heading)

	remaining := m.Height - headingHeight - 1
	sidebar := lipgloss.NewStyle().Width(m.Width * 1 / 4).Height(remaining).MarginLeft(2).Padding(1).Render(m.renderCryptos())
	body := lipgloss.NewStyle().Width((m.Width * 3 / 4) - 2).Height(remaining).Render("")

	
main := lipgloss.JoinHorizontal(
	lipgloss.Top,
	sidebar,
	body,
)

return bg.Render(
	lipgloss.JoinVertical(
		lipgloss.Top,
		heading,
		"\n",
		main,
	),
)
}

func(m *model)FetchLiveCryptos()tea.Cmd{
	return func() tea.Msg {
			req,err := http.NewRequest(http.MethodGet,fmt.Sprintf("%s/live-cryptos?limit=8",m.Url),nil)
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

	httpRes := AllCryptoResponse{}
	json.NewDecoder(resp.Body).Decode(&httpRes)


	return Cryptos(httpRes.Data)
	}
}

func(m *model)fetchCryptoByID()tea.Cmd{
	return func () tea.Msg {	
		req,err := http.NewRequest(http.MethodGet,fmt.Sprintf("https://api.alpstein.tech/api/v1/crypto/%s",m.CurrCryptoId),nil)
		if err != nil{
			log.Println(err)
		}
		req.Header.Set("Authorization",fmt.Sprintf("Bearer %s",m.Jwt))
		client := http.Client{}
		res,err := client.Do(req)
		if err != nil{
			log.Println(err)
		}
		defer res.Body.Close()
		
		var cryptoRes SingleCryptoResponse
		var CryptoMod CryptoModel
		if err := json.NewDecoder(res.Body).Decode(&cryptoRes); err != nil{
			log.Println(err)
		}
		json.Unmarshal(cryptoRes.Data[0],&CryptoMod)
		return SetCurrCrypto(CryptoMod)
	}
}

func (m *model) renderCryptos() string {
	var b strings.Builder

	for i, c := range m.Cryptos {
		rowStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.primaryTextColor))

		prefix := ""
		if i == m.Cursor {
			prefix = ">> "
			rowStyle = rowStyle.Foreground(
				lipgloss.Color(m.secondaryTextColor),
			)
		}

		// Symbol
		b.WriteString(rowStyle.Render(prefix + c.Symbol))
		b.WriteString("\n")

		// Heading truncation
		words := strings.Split(c.Heading, " ")
		max := 9
		if len(words) < max {
			max = len(words)
		}

		dots := ""
		if len(words) > 9 {
			dots = "..."
		}

		b.WriteString(
			rowStyle.Render(
				strings.Join(words[:max], " ") + dots,
			),
		)
		b.WriteString("\n\n")
	}

	return b.String()
}
