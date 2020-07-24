package app

import (
	"myhitbtcv4/model"
)

type MarginCalDBService struct{
	session *Session
}

func (u *MarginCalDBService) AddMar(mar *model.MarginDBVeh) error {
	u.session.marginCalDBChans.AddMarginDBChan <-mar
	marDbResp := <-mar.MCalDBRespChan
	return marDbResp.Err
}
func (u *MarginCalDBService) GetMar(mar *model.MarginDBVeh) (map[model.AppID]model.MarginParam, error) {
	u.session.marginCalDBChans.GetMarginDBChan <-mar
	resp := <-mar.MCalDBRespChan
	if resp.Err != nil{
		return nil, resp.Err
	}
	return resp.MarginDB, nil
}
func (u *MarginCalDBService) DeleteMar(mar *model.MarginDBVeh) error {
	user, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	if user.ID == mar.User.ID{
		u.session.marginCalDBChans.DeleteMarginDBChan <-mar
		marDbResp := <-mar.MCalDBRespChan
		return marDbResp.Err
	}
	return model.ErrUnauthorized
}