package app

import (
	"log"
	"myhitbtcv4/model"
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
		case dat = <-SessionDBChans.AddOrUpdateDbChan:
			db[dat.SessionID] = dat.Session
			if v, ok = db[dat.SessionID]; ok && v.ID == dat.SessionID {
				dat.CallerChan <- SessionDbResp{v.ID, v, nil}
			} else {
				dat.CallerChan <- SessionDbResp{Err: model.ErrInternal}
			}
		case dat = <-SessionDBChans.GetDbChan:
			if v, ok = db[dat.SessionID]; ok && v.ID == dat.SessionID {
				dat.CallerChan <- SessionDbResp{v.ID, v, nil}
			} else {
				log.Printf("v %v, ok %v", v, ok)
				dat.CallerChan <- SessionDbResp{Err: model.ErrInternal}
			}
		case dat = <-SessionDBChans.DeleteDbChan:
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
	session        Session
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
