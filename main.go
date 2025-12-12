package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/whiplashvin/alpstein-tui/loading"

	err "github.com/whiplashvin/alpstein-tui/error"

	"github.com/whiplashvin/alpstein-tui/dash"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)



type Screen int

const  (
	AuthScreen Screen = iota
	DashScreen
	LoadingScreen
	ErrorScreen

)
type userMsg string
type loadingState bool
type ErrorMessage string

type HttpResponse  struct {
	Message string 		`json:"message,omitempty"`
	Data 	interface{}	`json:"data,omitempty"`
}



type UserResType struct{
	Email string `json:"email"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	ImageUrl string `json:"imageUrl"`
}

type User struct {
	Message string   `json:"message"`
	Data    UserResType `json:"data"`
}

type model struct{
	BE_URL string
	OAUTH_CLIENT string
	OAUTH_CB string
	width int
	height int
	spinner  spinner.Model
	CurrUser string
	jwt interface{}
	Screen Screen
	t textinput.Model
	choices []string
	selected map[int]struct{}
	dashboard tea.Model
	loader tea.Model
	errorModel tea.Model
}

func initModel(u,c,cb string)*model{
	errMod := err.InitError()
	s := spinner.New()
	s.Spinner = spinner.Meter
	ti := textinput.New()
	ti.Placeholder = "auth-key"
	ti.Focus()
	ti.Cursor.BlinkSpeed = time.Millisecond * 500 
	ti.Cursor.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#74d4ff")) 
	ti.Width = 40
	ti.PlaceholderStyle = lipgloss.NewStyle().Background(lipgloss.Color("#18181b")).Foreground(lipgloss.Color("#555"))
	ti.TextStyle = lipgloss.NewStyle().Background(lipgloss.Color("#18181b")).Foreground(lipgloss.Color("#74d4ff")) 
	return &model{
		BE_URL: u,
		OAUTH_CLIENT: c,
		OAUTH_CB: cb,
		spinner: s,
		t: ti,
		choices: []string{"Code","Gym","Guitar"},
		selected: make(map[int]struct{}),
		errorModel:  errMod,
	}
}
func(m model)Init()tea.Cmd{
	m.generateSigninURL()
	return textinput.Blink
}
func(m model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	// var cmds []tea.Cmd
	switch msg := msg.(type){
		case err.ErrSignal:
				m.Screen = ErrorScreen
				var cmd tea.Cmd
				m.errorModel,cmd = m.errorModel.Update(msg)
				return m, cmd
		case dash.Cryptos:
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard,cmd = m.dashboard.Update(msg)
				return m,cmd 
			}
		case spinner.TickMsg:
    		if m.Screen == LoadingScreen {
        		var cmd tea.Cmd
        		m.loader, cmd = m.loader.Update(msg)
        		return m, cmd
    		}
		case tea.WindowSizeMsg:
        	m.width = msg.Width
			m.height = msg.Height
			var cmd tea.Cmd
			m.errorModel,cmd = m.errorModel.Update(msg)
        	return m, cmd
		 case userMsg:
        	m.CurrUser = string(msg)
        	dash := dash.InitDash(m.jwt,m.BE_URL,m.CurrUser,m.width,m.height)
        	m.dashboard = dash
			m.Screen = DashScreen
			return m, m.dashboard.Init()
		case loadingState:
			loader := loading.InitLoading(m.width,m.height)
			m.loader = loader
			m.Screen = LoadingScreen
        	return m, m.loader.Init()
		case ErrorMessage:
			cmd := func ()tea.Msg  {
				return err.ErrSignal{Msg: string(msg)}
			}
			return m, cmd
		case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c":
			return m,tea.Quit
		case "esc":
			switch m.Screen {
			case ErrorScreen:
				m.Screen = AuthScreen
			case DashScreen:
				m.Screen = AuthScreen	
			}
		case "enter":
			if m.Screen == AuthScreen {
       	 	   cmd := m.handleAuth()
        	   return m, cmd
    		}
		default: 	
			var cmd tea.Cmd
            m.t, cmd = m.t.Update(msg)
            return m, cmd
		}
	}
	var cmd tea.Cmd
	m.t, cmd = m.t.Update(msg)
    return m, cmd
	// return m,tea.Batch(cmds...)
}
func(m model)View()string{
	switch m.Screen {
	case AuthScreen:
		s := m.AuthScreen()
		return s
	case DashScreen:
		s := m.dashboard.View()
		return s
	case LoadingScreen:
		return m.loader.View()
	case ErrorScreen:
		s := m.errorModel.View()
		return s
	}
	return ""
}

func main(){
	godotenv.Load()
	url := os.Getenv("BACKEND_URL")
	oautClient := os.Getenv("OAUTH_CLIENT")
	oautCb := os.Getenv("OAUTH_CB")

	newModel := initModel(url,oautClient,oautCb)
	p := tea.NewProgram(*newModel,tea.WithAltScreen())
	p.Run()
}

func(m *model) handleAuth()tea.Cmd{
	auth := m.t.Value()
	temp := strings.TrimLeft(auth,"[")
	authKey := strings.TrimRight(temp,"]")
	res,errCmd := m.fetchJWT(authKey)
	if errCmd != nil{
		return errCmd
	}
	m.jwt = res
	return tea.Batch(
		m.handleLoading(true),
        m.getUserDetails(res),                        
	)
}

func(m *model) generateSigninURL(){
	   var (
    	googleOAuthConfig = &oauth2.Config{
    		ClientID:     m.OAUTH_CLIENT, 
    		RedirectURL:  m.OAUTH_CB, 
    		Scopes:       []string{"profile", "email"}, 
    		Endpoint:     google.Endpoint,
    	}
		oauthStateString = "CLI-app" 
    )
	url := googleOAuthConfig.AuthCodeURL(oauthStateString)
	fmt.Println(url)
	// openBrowser(url)    
}

func openBrowser(url string) error {
        cmd := "open"
        args := []string{url}
    return exec.Command(cmd, args...).Start()
}

func (m *model)fetchJWT(authKey string)(interface{},tea.Cmd){
	resp ,err := http.Get(fmt.Sprintf("%scli-oauth/callback?auth-key=%s",m.BE_URL,authKey))
	if err != nil{
		return "",m.handleError(err.Error())
	}

	defer resp.Body.Close()
	body := resp.Body

	httpRes := HttpResponse{}
	if err := json.NewDecoder(body).Decode(&httpRes); err != nil{
		return "",m.handleError(err.Error())
	}

	if resp.StatusCode == http.StatusNotFound{
		return "",m.handleError(httpRes.Message)
	}
	return httpRes.Data,nil
}
func(m *model) getUserDetails(jwt interface{})tea.Cmd{
	return func () tea.Msg {	
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/user",m.BE_URL), nil)
		if err != nil {
			return m.handleError(err.Error())
		}
		req.Header.Add("Authorization",fmt.Sprintf("Bearer %s",jwt))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		
		if err != nil{
			return m.handleError(err.Error())
		}
		defer resp.Body.Close()
		
	
		user := User{}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil{
				return m.handleError(err.Error())
		}
		if resp.StatusCode == http.StatusNotFound {
			return m.handleError(user.Message)
		}
		return userMsg(user.Data.FirstName)
	}
}
func (m *model)handleLoading(state bool)tea.Cmd{
	return func() tea.Msg {
		return loadingState(state)
	}
}
func (m *model)handleError(msg string)tea.Cmd{
	return func() tea.Msg {
		return ErrorMessage(msg)
	}
}

func (m *model) AuthScreen() string {
    bg := lipgloss.NewStyle().
        Width(m.width).
        Height(m.height).
        Background(lipgloss.Color("#18181b"))

    // Title
    titleStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("#18181b")).
        Foreground(lipgloss.Color("#74d4ff")).
        AlignHorizontal(lipgloss.Center).
        Width(m.width).PaddingTop(2)

    title := titleStyle.Render("Welcome to Alpstein TUI")
    titleHeight := lipgloss.Height(title) + 1

    // Message
    message := "Please paste the auth key and press enter."
    messageCentered := lipgloss.NewStyle().Background(lipgloss.Color("#18181b")).Foreground(lipgloss.Color("#74d4ff")).Width(m.width).AlignHorizontal(lipgloss.Center).Render(message)
    messageHeight := 1

    // Input field
	inputStyle := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Center).Background(lipgloss.Color("#18181b"))
	containerStyle := lipgloss.NewStyle().
    Width(45).
    AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Color("#18181b")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#74d4ff")).Padding(0,1)
	temp := containerStyle.Render(m.t.View())
    centeredinput := inputStyle.Render(temp)
    inputHeight := lipgloss.Height(temp)

    // Total content height (message + small gap + input)
    gap := 1
    blockHeight := messageHeight + gap + inputHeight

    // Remaining space after title
    remaining := m.height - titleHeight
    if remaining < blockHeight {
        remaining = blockHeight
    }

    topPad := (remaining - blockHeight) / 2
    bottomPad := remaining - blockHeight - topPad

    var b strings.Builder

    b.WriteString(title)
    b.WriteString("\n")
    b.WriteString(strings.Repeat("\n", topPad))
    b.WriteString(messageCentered)
    b.WriteString("\n\n") // message → input gap
    b.WriteString(centeredinput)
    b.WriteString(strings.Repeat("\n", bottomPad))

    return bg.Render(b.String())
}

// cmj2mum88000901o848ygdwzg