package model

import (
	"time"
	"net/http"
	jwt "github.com/dgrijalva/jwt-go"	
)
type SessionID string

//Session carries the state parameters of a particular user
type Session struct {
	ID            SessionID
	UsrID         UserID
	ApID          AppID
	CachedUser    *User
	Sockets        map[string]AppConn
	//cachedApp     *App
	AppMemDBChans    MDDBChans
	AppMemDBService    AppMemDBServicer
	AppBoltDBChans    ABDBChans
	AppBoltDBService    AppBoltDBServicer 
	UserBoltDBChans   UDBChans
	UserBoltDBService   UserBoltDBServicer
	MarginCalDBChans  MCalDBChans
	MarginCalDBService   MarginCalDBServicer
	WorkerAppService WorkerAppServicer
	WebsocketUserService WebsocketUserServicer
	WebsocketUSChans WUSChans
}
// JWT schema of the data it will store
type claims struct {
	Username    string `json:"username"`
	ID          UserID `json:"id"`
	RedirectURL string `json:"redirecturl"`
	Level       string `json:"level"`
	Token       string `json:"-"`
	jwt.StandardClaims
}

func (s *Session) SetToken(w http.ResponseWriter, r *http.Request, username string, userid UserID, redirectURL, level string) {
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
func (s *Session) Authenticate() (*User, error) {
	// Return user if already authenticated.
	if s.CachedUser.Username != "" {
		return s.CachedUser, nil
	}
	return nil, ErrUserNotCached
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

type SessioneDBServicer interface{
	AddSession(session *Session) error
	GetSession(id SessionID) (*Session, error)
	UpdateSession(session *Session) error
	DeleteSession(id SessionID) error
	AlreadyLoggedIn(w http.ResponseWriter, r *http.Request) (*User, bool)
}


