package app

import (
	"log"
	"myhitbtcv4/model"
)

type MarginCalDBService struct{
	session *Session
}
func (u *MarginCalDBService) GetMargin(mar model.MarginDBVeh) (map[model.AppID]model.MarginParam, error) {
	MarginDB := make(map[model.AppID]model.MarginParam)
	for k,v := range mar.User.ApIDs{
		app, err := u.session.appMemDBService.GetApp(v)
		if err != nil {
			log.Printf("GetMar: %s Error Geting App for Margin proces: Error %v", k, err)
		}
		MarginDB[app.Data.ID] =  <-app.Chans.MParamChan
	}
	if len(MarginDB) != 0 {
		return MarginDB, nil
	} else{
		return MarginDB, model.ErrAppNotFound
	}
}