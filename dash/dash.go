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
	"github.com/gorilla/websocket"
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
type WSMsg struct{
	Event string `json:"event"`
	Payload string `json:"payload"`
}
type WSResp struct{
	Kind string `json:"kind"`
	Value float64 `json:"value"`
}
type CryptoQueryMetadata struct{
	HasNextPage   bool   `json:"hasNextPage"`
	HasPrevPage   bool   `json:"hasPrevPage"`
	LastSeenTime  int64  `json:"lastSeenTime"`
	FirstSeenTime int64  `json:"firstSeenTime"`
	LastSeenId    string `json:"lastSeenId"`
	FirstSeenId   string `json:"firstSeenId"`
}
type AllCryptoResponse struct{
	Data 	 []CryptoModel 		 `json:"data"`
	Metadata CryptoQueryMetadata `json:"metadata"`
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
	tertiaryTextColor string
	borderColor string
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
	PositionDisplayed string
	QueryMetada CryptoQueryMetadata
	debounceID int
	WSConn *websocket.Conn
	WSChan chan WSResp
	WSRes WSResp
}

type DebounceFetch struct {
	id int
}
type LiveCryptosLoaded struct {
	Cryptos  []CryptoModel
	Metadata CryptoQueryMetadata
}
type Cryptos []CryptoModel
type QueryMetada CryptoQueryMetadata 
type SetCryptoId string
type SetCurrCrypto CryptoModel
type WSConnected struct {
	Conn *websocket.Conn
}
type WSRespSingnal WSResp
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
		tertiaryTextColor:	"#52525c",
		borderColor: "#27272a",
		Width: width,
		Height: height,
		Jwt: jwt,
		Url: url,
		CurrUser: currUser,
		PositionDisplayed: "long",
	}
}

func (m model)Init()tea.Cmd{
	return tea.Batch(m.FetchLiveCryptos())
}

func (m model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
	case PositionDisplayed:
		m.PositionDisplayed = string(msg)
	case DebounceFetch:
		if msg.id != m.debounceID {
			return m, nil
		}
		return m, func() tea.Msg {
			return SetCryptoId(m.Cryptos[m.Cursor].Id)
		}
	case WSConnected:
		if m.WSConn != nil {
			m.WSConn.Close()
		}
		m.WSConn = msg.Conn
		return m, m.readFromWSS()
	case WSRespSingnal:
		m.WSRes.Kind = msg.Kind
		m.WSRes.Value = msg.Value
		return m, m.readFromWSS() // 🔁 KEEP LISTENING
	case SetCryptoId:
		m.CurrCryptoId = string(msg)
		cmd := m.fetchCryptoByID()
		return m, cmd
	case SetCurrCrypto:	
		m.CurrCrypto = CryptoModel(msg)
		if msg.Position == "unclear"{
			m.PositionDisplayed = "long"
		}else{
			m.PositionDisplayed = msg.Position
		}
		cmd := m.connectToWS()
		return m,cmd
	case LiveCryptosLoaded:
		m.Cryptos = msg.Cryptos
		m.CurrCryptoId = m.Cryptos[0].Id
		m.CurrCrypto = m.Cryptos[0]
		m.QueryMetada = msg.Metadata
		m.Cursor = 0
		cmd := m.connectToWS()
		return m,cmd
	case tea.WindowSizeMsg:
        m.Width = msg.Width
		m.Height = msg.Height
        return m, nil
	case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c":
			if m.WSConn != nil{
				m.WSConn.Close()
			}
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
		case "n":
			var cmd tea.Cmd
			if m.QueryMetada.HasNextPage{
				cmd = m.FetchNextCryptoBatch()
				return m,cmd
			}
		case "p":
			var cmd tea.Cmd
			if m.QueryMetada.HasPrevPage{
				cmd = m.FetchPrevCryptoBatch()
				return m,cmd
			}
		case "s":
			var cmd tea.Cmd
			if m.CurrCrypto.Position == "unclear"{
				cmd = func ()tea.Msg  {
					return PositionDisplayed("short")
				}
				return m,cmd
			}
		case "l":
			var cmd tea.Cmd
			if m.CurrCrypto.Position == "unclear"{
				cmd = func ()tea.Msg  {
					return PositionDisplayed("long")
				}
				return m,cmd
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
	Render(fmt.Sprintf("hey %s! 🫡", m.CurrUser))

	heading := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)

	headingHeight := lipgloss.Height(heading)

	gap := 1 + 10
	remaining := m.Height - headingHeight - gap
	sidebar := lipgloss.NewStyle().Width((m.Width * 1 / 4) - 2).Height(remaining).MarginLeft(2).Padding(1).
	Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(m.borderColor)).
	Render(m.renderCryptos())
	body := lipgloss.NewStyle().Width((m.Width * 3 / 4) - 4).Height(remaining).AlignHorizontal(lipgloss.Center).
	Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(m.borderColor)).
	Render(m.renderCryptoyID())

	
main := lipgloss.JoinHorizontal(
	lipgloss.Top,
	sidebar,
	body,
)

var footerStinng string
if m.QueryMetada.HasPrevPage{
	footerStinng += "[p] prev "
}
if m.QueryMetada.HasNextPage {
	footerStinng += "[n] next "
}
footerStinng += "[d] docs "
footerStinng += "[t] trades "
footerStinng += "[▲] up "
footerStinng += "[▼] down "
footer := lipgloss.NewStyle().Width(m.Width-2).Height(1).Foreground(lipgloss.Color(m.tertiaryTextColor)).
MarginLeft(2).AlignHorizontal(lipgloss.Center).Render(footerStinng)
return bg.Render(
	lipgloss.JoinVertical(
		lipgloss.Top,
		heading,
		"\n\n\n",
		main,
		"\n\n\n",
		footer,
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

	return LiveCryptosLoaded{
		Cryptos: httpRes.Data,
		Metadata: httpRes.Metadata,
		}
	}
}
func (m *model)FetchNextCryptoBatch()tea.Cmd{
	return func() tea.Msg {
		req,err := http.NewRequest(http.MethodGet,fmt.Sprintf("%s/live-cryptos?action=next&limit=8&last_seen=%d|%s",m.Url,m.QueryMetada.LastSeenTime,m.QueryMetada.LastSeenId),nil)
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

		return LiveCryptosLoaded{
			Cryptos: httpRes.Data,
			Metadata: httpRes.Metadata,
		}
	}
}
func (m *model)FetchPrevCryptoBatch()tea.Cmd{
	return func() tea.Msg {
		req,err := http.NewRequest(http.MethodGet,fmt.Sprintf("%s/live-cryptos?action=prev&limit=8&last_seen=%d|%s",m.Url,m.QueryMetada.FirstSeenTime,m.QueryMetada.FirstSeenId),nil)
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

		log.Println(httpRes)
		return LiveCryptosLoaded{
			Cryptos: httpRes.Data,
			Metadata: httpRes.Metadata,
		}
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

func (m *model)renderCryptos()string{
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

		x := ""
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00d3f2"))
		if c.Status == "triggered"{
			x += statusStyle.Render("⦿")
		}
		// Symbol
		b.WriteString(rowStyle.Render(prefix + c.Symbol + " " +x))
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

func(m *model)renderCryptoyID()string{
	if m.CurrCrypto.Id != "" {
		var s  = lipgloss.NewStyle().Foreground(lipgloss.Color(m.secondaryTextColor)).AlignHorizontal(lipgloss.Center).MarginTop(2)
		heading := m.CurrCrypto.Heading
		heading += m.CurrCrypto.Name
		
		agentsOp := "Agent's Opinion \n"
		agentsOp += m.renderSignals()

		trivia := ""
		if m.CurrCrypto.Position == "unclear"{
			trivia += "[s] short position [l] long position"
		}

		liveStats := "Live Stats ⚡️\n"
		liveStats += m.renderLiveStats()

		var output strings.Builder
		output.WriteString(heading)
		output.WriteString("\n\n\n")
		output.WriteString(agentsOp)
		output.WriteString("\n")
		output.WriteString(trivia)
		output.WriteString("\n\n\n")
		output.WriteString(liveStats)
		return s.Render(output.String())
	}else {
		return ""
	}
}
type PositionDisplayed string

func (m *model) renderSignals() string {
	box := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.secondaryTextColor)).
		AlignHorizontal(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.tertiaryTextColor)).
		Padding(0, 1).Width(15)

	position := box.Render(
		"Position\n" + m.CurrCrypto.Position,
	)

	var action = ""
	switch m.PositionDisplayed{
	case "long":
		action +=  box.Render("Buy\n" + fmt.Sprintf("%.2f", m.CurrCrypto.BuyPrice))
	case "short":
		action += box.Render("Sell\n" + fmt.Sprintf("%.2f", m.CurrCrypto.SellPrice))
	}

	var tp = ""
	switch m.PositionDisplayed{
	case "long":
		tp += box.Render("Take Profit\n" + fmt.Sprintf("%.2f", m.CurrCrypto.TakeProfit))
	case "short":
		tp += box.Render("Take Profit\n" + fmt.Sprintf("%.2f", m.CurrCrypto.ShortCoverProfit))
	}

	var sl = ""
	switch m.PositionDisplayed{
	case "long":
		sl += box.Render("Stop Loss\n" + fmt.Sprintf("%.2f", m.CurrCrypto.StopLoss))
	case "short":
		sl += box.Render("Stop Loss\n" + fmt.Sprintf("%.2f", m.CurrCrypto.ShortCoverLoss))
	}

	trivia := ""
	if m.CurrCrypto.Position == "unclear"{
		trivia += ""
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Top, 
		position,
	    action,
		tp,
		sl,
	)
}
func (m *model)renderLiveStats()string{
	box := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.secondaryTextColor)).
		AlignHorizontal(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.tertiaryTextColor)).
		Padding(0, 1).Width(20)

	creation := ""
	creation += box.Render("Creation price\n" + fmt.Sprintf("%.2f",m.CurrCrypto.PriceAtCreation))

	status := ""
	status += box.Render("Status\n" + m.CurrCrypto.Status)

	trigPos := ""
	trigPos += box.Render("Triggered\n" + m.CurrCrypto.TriggeredPosition)

	s := "P&L" + "\n"
	if m.CurrCrypto.Status == "triggered"{
		if m.WSRes.Kind == "loss"{
			box = box.Foreground(lipgloss.Color("#fb2c36"))
			s += "-" + " "
			}else{
				box = box.Foreground(lipgloss.Color("#00c950"))
				s += "+" + " "
			}
		s += fmt.Sprintf("%.2f",m.WSRes.Value)
	}else{
		s += "-"
	}
	pl := box.Render(s)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		creation,
		status,
		trigPos,
		pl,
	)
}

func (m *model) connectToWS() tea.Cmd {
	return func() tea.Msg {
		dialer := websocket.Dialer{
			Proxy: http.ProxyFromEnvironment,
		}

		headers := http.Header{}
		headers.Set("Origin", "https://alpstein.tech")

		conn, _, err := dialer.Dial("wss://ws.alpstein.tech", headers)
		if err != nil {
			log.Println("WS dial error:", err)
			return nil
		}

		// send SUB
		_ = conn.WriteJSON(WSMsg{
			Event:   "SUB",
			Payload: m.CurrCryptoId,
		})

		return WSConnected{Conn: conn}
	}
}

func (m *model)readFromWSS()tea.Cmd{
	return func() tea.Msg {
		if m.WSConn == nil {
			return nil
		}
		_, p, err := m.WSConn.ReadMessage()
		if err != nil {
			log.Println("WS read error:", err)
			return nil
		}

		var resp WSResp
		if err := json.Unmarshal(p, &resp); err != nil {
			return nil
		}

		return WSRespSingnal(resp)
	}
}

