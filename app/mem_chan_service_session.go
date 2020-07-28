package app

import (
	"log"
	"myhitbtcv4/model"
	"net/http"
	//"sync"
	"time"
)

type SDBChans struct {
	AddOrUpdateDbChan chan SessionDbData
	GetDbChan         chan SessionDbData
	DeleteDbChan      chan SessionDbData
}

type SessionDB map[model.SessionID]*Session

func SessionMemDBServiceFunc(SessionMemDBChans SDBChans) {
	var (
		dat SessionDbData
		v   *Session
		ok  bool
	)
	db := make(SessionDB) //the mainDataBase very impotant
	for {
		select {
		case dat = <-SessionMemDBChans.AddOrUpdateDbChan:
			db[dat.SessionID] = dat.Session
			if v, ok = db[dat.SessionID]; ok && v.ID == dat.SessionID {
				dat.CallerChan <- SessionDbResp{v.ID, v, nil}
			} else {
				dat.CallerChan <- SessionDbResp{Err: model.ErrInternal}
			}
		case dat = <-SessionMemDBChans.GetDbChan:
			if v, ok = db[dat.SessionID]; ok && v.ID == dat.SessionID {
				dat.CallerChan <- SessionDbResp{v.ID, v, nil}
			} else {
				log.Printf("v %v, ok %v", v, ok)
				dat.CallerChan <- SessionDbResp{Err: model.ErrInternal}
			}
		case dat = <-SessionMemDBChans.DeleteDbChan:
			delete(db, dat.SessionID)
			if v, ok = db[dat.SessionID]; ok && v.ID == dat.SessionID {
				dat.CallerChan <- SessionDbResp{Err: model.ErrInternal}
			} else {
				dat.CallerChan <- SessionDbResp{}
			}
		}
	}
}

type SessionDBService struct {
	sessionDBChans SDBChans
	session        model.Session
}

func NewSession(uchans model.UDBChans, abdbchans model.ABDBChans, mdchans MDDBChans, wuschans WUSChans, mcalchans model.MCalDBChans, schans SDBChans, sess Session) *SessionDBService {
	s := model.Session{
		UserBoltDBChans: uchans,
		AppBoltDBChans: abdbchans,
		AppMemDBChans:  mdchans,
		WebsocketUSChans: wuschans,		
		MarginCalDBChans: mcalchans, 
	}
	s.UserBoltDBService.session = &s
	s.AppBoltDBService.session = &s
	s.MarginCalDBService.session = &s
	return &SessionDBService{
		sessionDBChans: schans,
		session:        s,
	}
}

func (u *SessionDBService) AddSession(session *Session) error {
	CallerChan := make(chan SessionDbResp)
	u.sessionDBChans.AddOrUpdateDbChan <- SessionDbData{session.ID, session, CallerChan}
	sessDbResp := <-CallerChan
	return sessDbResp.Err
}
func (u *SessionDBService) GetSession(id model.SessionID) (*Session, error) {
	if id == "" {
		return nil, model.ErrSessionNameEmpty
	}
	CallerChan := make(chan SessionDbResp)
	u.sessionDBChans.GetDbChan <- SessionDbData{id, nil, CallerChan}
	sessDbResp := <-CallerChan
	if sessDbResp.Session != nil && sessDbResp.Err == nil {
		return sessDbResp.Session, nil
	}
	return sessDbResp.Session, sessDbResp.Err
}
func (u *SessionDBService) UpdateSession(session *Session) error {
	cachedUser, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update session.
	CallerChan := make(chan SessionDbResp)
	u.sessionDBChans.GetDbChan <- SessionDbData{session.ID, nil, CallerChan}
	sessDbResp := <-CallerChan
	if sessDbResp.Session != nil && sessDbResp.Err == nil {
		if sessDbResp.Session.ID != cachedUser.SessID {
			return model.ErrUnauthorized
		}
		u.sessionDBChans.AddOrUpdateDbChan <- SessionDbData{session.ID, session, CallerChan}
		sessDbResp := <-CallerChan
		return sessDbResp.Err
	}
	return model.ErrSessionNotFound
}
func (u *SessionDBService) DeleteSession(id model.SessionID) error {
	cachedSession, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update session.
	CallerChan := make(chan SessionDbResp)
	u.sessionDBChans.GetDbChan <- SessionDbData{id, nil, CallerChan}
	sessDbResp := <-CallerChan
	if sessDbResp.Session != nil && sessDbResp.Err == nil {
		if sessDbResp.Session.ID != cachedSession.SessID {
			return model.ErrUnauthorized
		}
		u.sessionDBChans.DeleteDbChan <- SessionDbData{id, nil, CallerChan}
		sessDbResp := <-CallerChan
		return sessDbResp.Err
	}
	return model.ErrSessionNotFound
}
func (u *SessionDBService) Authenticate() (*model.User, error) {
	// Return user if already authenticated.
	return u.session.Authenticate()
}
// deletes the cookie
func (u *SessionDBService) logout(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("Auth")
	if err != nil {
		return
	}
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
	http.SetCookie(w, &deleteCookie)
	return
}
//AlreadyLoggedIn is use to ensure user are properly authenticated before having access to handler resources
func (u *SessionDBService) AlreadyLoggedIn(w http.ResponseWriter, r *http.Request) (*model.User, bool) {
	cookie, err := r.Cookie("Auth")
	if err != nil {
		return nil, false
	}
	token, err := jwt.ParseWithClaims(cookie.Value, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("sEcrEtPassWord!234"), nil
	})
	if err != nil {
		return nil, false
	}
	if userClaims, ok := token.Claims.(*claims); ok && token.Valid {
		user, err := u.session.userBoltDBService.GetUserByName(userClaims.Username)
		if err == nil {
			userSession, err := u.GetSession(user.SessID)
			if err != nil {
				log.Printf("AlreadyLoggedIn1 %v : SessID = %v user = %v\n", err, user.SessID, user)
				panic("Unable to get User Session from DB")
			}
			u.session = *userSession
			u.session.cachedUser = user
			return user, true
		}
	}
	return nil, false
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

