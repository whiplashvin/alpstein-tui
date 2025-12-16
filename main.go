package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/whiplashvin/alpstein-tui/loading"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	err "github.com/whiplashvin/alpstein-tui/error"

	"github.com/whiplashvin/alpstein-tui/dash"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.7"

type Screen int

const  (
	AuthScreen Screen = iota
	DashScreen
	LoadingScreen
	ErrorScreen

)
type userMsg string
type ErrorMessage string
type startAuthMsg struct{}

type jwtResultMsg struct {
	jwt string
	err string
}

type HttpResponse  struct {
	Message string 		`json:"message,omitempty"`
	Data 	string	`json:"data,omitempty"`
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
	CurrUser string
	jwt string
	Screen Screen
	t textinput.Model
	dashboard tea.Model
	loader tea.Model
	errorModel tea.Model
	bgColor string
	primaryTextColor string
	secondaryTextColor string
}

func initModel(u,c,cb string)*model{
	loaderMod := loading.InitLoading()
	errMod := err.InitError()
	ti := textinput.New()
	ti.Placeholder = "auth-key"
	ti.Focus()
	ti.Cursor.BlinkSpeed = time.Millisecond * 500 
	ti.Cursor.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#a3b3ff")) 
	ti.Width = 40
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a3b3ff")) 
	return &model{
		BE_URL: u,
		OAUTH_CLIENT: c,
		OAUTH_CB: cb,
		t: ti,
		errorModel:  errMod,
		loader: loaderMod,
		// bgColor: "#18181a",
		primaryTextColor: "#a3b3ff",
		secondaryTextColor: "#c7d8ff",
	}
}
func(m model)Init()tea.Cmd{
	log.Println("Program started")
	m.generateSigninURL()
	return tea.Batch(textinput.Blink,m.loader.Init())
}
func(m model)Update(msg tea.Msg)(tea.Model,tea.Cmd){
	switch msg := msg.(type){
		case dash.PositionDisplayed:
			var cmd tea.Cmd
			m.dashboard,cmd = m.dashboard.Update(msg)
			return m,cmd
		case dash.WSConnected:
			var cmd tea.Cmd
			m.dashboard,cmd = m.dashboard.Update(msg)
			return m,cmd
		case dash.WSRespSingnal:
			var cmd tea.Cmd
			m.dashboard,cmd = m.dashboard.Update(msg)
			return m,cmd
		case dash.DebounceFetch:
			var cmd tea.Cmd
			m.dashboard,cmd = m.dashboard.Update(msg)
			return m,cmd
		case dash.SetCryptoId:
			var cmd tea.Cmd
			m.dashboard,cmd = m.dashboard.Update(msg)
			return m,cmd
		case dash.SetCurrCrypto:	
			var cmd tea.Cmd
			m.dashboard,cmd = m.dashboard.Update(msg)
			return m,cmd
		case jwtResultMsg:
			if msg.err != "" {
				return m, m.handleError(msg.err)
			}
			m.jwt = msg.jwt
		return m, m.getUserDetails(msg.jwt)
		case startAuthMsg:
    		return m, m.handleAuth()
		case err.ErrSignal:
				m.Screen = ErrorScreen
				var cmd tea.Cmd
				m.errorModel,cmd = m.errorModel.Update(msg)
				return m, cmd
		case dash.LiveCryptosLoaded:
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard,cmd = m.dashboard.Update(msg)
				return m,cmd 
			}
		case spinner.TickMsg:
        		var cmd tea.Cmd
        		m.loader, cmd = m.loader.Update(msg)
        		return m, cmd
		case tea.WindowSizeMsg:
        	m.width = msg.Width
			m.height = msg.Height
			var cmd tea.Cmd
			var cmd1 tea.Cmd
			// var cmd2 tea.Cmd
			m.errorModel,cmd = m.errorModel.Update(msg)
			m.loader,cmd1 = m.loader.Update(msg)
			// m.dashboard,cmd2 = m.dashboard.Update(msg)
        	return m, tea.Batch(cmd,cmd1)
		 case userMsg:
        	m.CurrUser = string(msg)
        	dash := dash.InitDash(m.jwt,m.BE_URL,m.CurrUser,m.width,m.height)
        	m.dashboard = dash
			m.Screen = DashScreen
			return m, m.dashboard.Init()
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
				m.Screen = LoadingScreen
				return m, 
				func() tea.Msg {
            return startAuthMsg{}
        }
    		}
		case "down":
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard, cmd = m.dashboard.Update(msg)
				return m, cmd
			}
		case "up":
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard, cmd = m.dashboard.Update(msg)
				return m, cmd
			}
		case "n":
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard, cmd = m.dashboard.Update(msg)
				return m, cmd
			}
		case "p":
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard, cmd = m.dashboard.Update(msg)
				return m, cmd
			}
		case "s":
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard, cmd = m.dashboard.Update(msg)
				return m, cmd
			}
		case "l":
			if m.Screen == DashScreen{
				var cmd tea.Cmd
				m.dashboard, cmd = m.dashboard.Update(msg)
				return m, cmd
			}
		}
	}
	var cmd tea.Cmd
	m.t, cmd = m.t.Update(msg)
    return m, cmd
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
	f, _ := os.Create("alpstein.log")
    log.SetOutput(f)
	godotenv.Load()
	url := os.Getenv("BACKEND_URL")
	oautClient := os.Getenv("OAUTH_CLIENT")
	oautCb := os.Getenv("OAUTH_CB")


	showVersion := flag.Bool("version", false, "print version and exit")
    flag.Parse()
    if *showVersion {
        fmt.Println(version)
        os.Exit(0)
    }

	newModel := initModel(url,oautClient,oautCb)
	p := tea.NewProgram(*newModel,tea.WithAltScreen(),tea.WithMouseCellMotion())
	p.Run()
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
	openBrowser(url)    
}

func openBrowser(url string) error {
        cmd := "open"
        args := []string{url}
    return exec.Command(cmd, args...).Start()
}

func (m *model) handleAuth() tea.Cmd {
	auth := strings.Trim(m.t.Value(), "[]")

	return func() tea.Msg {
		resp, err := http.Get(
			fmt.Sprintf("%scli-oauth/callback?auth-key=%s", m.BE_URL, auth),
		)
		if err != nil {
			return jwtResultMsg{err: err.Error()}
		}
		defer resp.Body.Close()

		var httpRes HttpResponse
		if err := json.NewDecoder(resp.Body).Decode(&httpRes); err != nil {
			return jwtResultMsg{err: err.Error()}
		}

		if resp.StatusCode == http.StatusNotFound {
			return jwtResultMsg{err: httpRes.Message}
		}

		return jwtResultMsg{jwt: httpRes.Data}
	}
}

func(m *model) getUserDetails(jwt interface{})tea.Cmd{
	return func () tea.Msg {	
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/user",m.BE_URL), nil)
		if err != nil {
			// return m.handleError(err.Error())
			return ErrorMessage(err.Error())
		}
		req.Header.Add("Authorization",fmt.Sprintf("Bearer %s",jwt))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		
		if err != nil{
			// return m.handleError(err.Error())
			return ErrorMessage(err.Error())
		}
		defer resp.Body.Close()
		
	
		user := User{}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil{
				return m.handleError(err.Error())
		}
		if resp.StatusCode == http.StatusNotFound {
			// return m.handleError(user.Message)
			return ErrorMessage(user.Message)
		}
		return userMsg(user.Data.FirstName)
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
        Background(lipgloss.Color(m.bgColor)).Margin(0).Padding(0)

    logoStyle := lipgloss.NewStyle().
        Background(lipgloss.Color(m.bgColor)).
        Foreground(lipgloss.Color(m.primaryTextColor)).
        AlignHorizontal(lipgloss.Center).
        Width(m.width)

    logo := logoStyle.Render(`
	
--         _      _                 _            _         		-- 
--        / \    | |  _ __    ___  | |_    ___  (_)  _ __  		-- 
--       / _ \   | | | '_ \  / __| | __|  / _ \ | | | '_ \ 		-- 
--      / ___ \  | | | |_) | \__ \ | |_  |  __/ | | | | | |		-- 
--     /_/   \_\ |_| | .__/  |___/  \__|  \___| |_| |_| |_|		-- 
--                   |_|                                   		-- 

`)
    logoHeight := lipgloss.Height(logo) + 1

	brandingStyle := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Color(m.bgColor)).Foreground(lipgloss.Color(m.primaryTextColor)).Bold(true).Underline(true)
	branding := brandingStyle.Render("AI-powered real-time crypto insights. Sans noise.")

	subBrandingStyle := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Color(m.bgColor)).Foreground(lipgloss.Color(m.secondaryTextColor))
	subBranding := subBrandingStyle.Render("Stay ahead with intelligent crypto analysis. Intense, data heavy blogs cleansed and made actionable.")

    message := "enter the auth key and hit enter"
    messageCentered := lipgloss.NewStyle().Background(lipgloss.Color(m.bgColor)).Foreground(lipgloss.Color(m.primaryTextColor)).Width(m.width).AlignHorizontal(lipgloss.Center).Render(message)
    messageHeight := 1

    // Input field
	inputStyle := lipgloss.NewStyle().Width(m.width).AlignHorizontal(lipgloss.Center).Background(lipgloss.Color(m.bgColor))
	containerStyle := lipgloss.NewStyle().
    Width(45).
    AlignHorizontal(lipgloss.Center).
	Background(lipgloss.Color(m.bgColor)).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(m.primaryTextColor)).Padding(0,1)
	temp := containerStyle.Render(m.t.View())
    centeredinput := inputStyle.Render(temp)
    inputHeight := lipgloss.Height(temp)

    // Total content height (message + small gap + input)
    gap := 1
    blockHeight := messageHeight + gap + inputHeight

    // Remaining space after logo
    remaining := m.height - logoHeight
    if remaining < blockHeight {
        remaining = blockHeight
    }

    topPad := (remaining - blockHeight) / 2
    bottomPad := remaining - blockHeight - topPad

    var b strings.Builder

    b.WriteString(logo)
    b.WriteString("\n")
	b.WriteString(branding)
	b.WriteString("\n\n")
	b.WriteString(subBranding)
    b.WriteString(strings.Repeat("\n", topPad))
    b.WriteString(messageCentered)
    b.WriteString("\n\n") // message → input gap
    b.WriteString(centeredinput)
    b.WriteString(strings.Repeat("\n", bottomPad))

    return bg.Render(b.String())
}