package app

import (
	"myhitbtcv4/model"
	"net/http"
	//"sync"
	"time"
	"github.com/gorilla/websocket"
	jwt "github.com/dgrijalva/jwt-go"
)
type AppConn struct{
	Conn *websocket.Conn
	CloseAppSocketChan chan bool
}
//Session carries the state parameters of a particular user
type Session struct {
	ID            model.SessionID
	UsrID         model.UserID
	ApID          model.AppID
	cachedUser    *model.User
	Sockets        map[string]AppConn
	cachedApp     *App
	appMemDBChans    MDDBChans
	appMemDBService    AppMemDBService
	appBoltDBChans    MDDBChans
	appBoltDBService    AppMemDBService
	userBoltDBChans   model.UDBChans
	userBoltDBService   UserBoltDBService
	workerAppService WorkerAppService
	websocketUserService WebsocketUserService
	websocketUSChans WUSChans
}

func NewSession(uchans model.UDBChans, mdchans MDDBChans, wuschans WUSChans) Session {
	s := Session{
		userDBChans: uchans,
		appDBChans:  mdchans,
		websocketUSChans: wuschans,
	}
	s.userDBService.session = &s
	return s
}

// JWT schema of the data it will store
type claims struct {
	Username    string `json:"username"`
	ID          model.UserID `json:"id"`
	RedirectURL string `json:"redirecturl"`
	Level       string `json:"level"`
	Token       string `json:"-"`
	jwt.StandardClaims
}

// create a JWT and put in the clients cookie
func (s *Session) SetToken(w http.ResponseWriter, r *http.Request, username string, userid model.UserID, redirectURL, level string) {
	expireCookie := time.Now().Add(time.Minute * 20)
	expireToken := time.Now().Add(time.Minute * 20).Unix()
	userClaims := &claims{
		username,
		userid,
		redirectURL,
		level,
		"tokenkey",
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "localhost:8080",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	signedToken, _ := token.SignedString([]byte("sEcrEtPassWord!234"))
	cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true}
	http.SetCookie(w, &cookie)
	time.Sleep(300 * time.Millisecond)
	// redirect
	http.Redirect(w, r, userClaims.RedirectURL, http.StatusSeeOther)
	return
}
func (s *Session) Authenticate() (*model.User, error) {
	// Return user if already authenticated.
	if s.cachedUser.Username != "" {
		return s.cachedUser, nil
	}
	return nil, model.ErrUserNotCached
}
// deletes the cookie
func (s *Session) logout(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("Auth")
	if err != nil {
		return
	}
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
	http.SetCookie(w, &deleteCookie)
	return
}

type autoTradeData struct {
	Host      string
	PublicKey string
	Secret    string
	SymID     string
}

type dashboardTotalData struct {
	totalBalance float64
	totalProfit  float64
}
type dashboardSymData struct {
	symBaseBalance  float64
	symQuoteBalance float64
	symBaseProfit   float64
	symQuoteProfit  float64
	dateStarted     time.Time
	id              string
}
type SessionDbData struct {
	SessionID  model.SessionID
	Session    *Session
	CallerChan chan SessionDbResp
}

type SessionDbResp struct {
	SessionID model.SessionID
	Session   *Session
	Err       error
}
