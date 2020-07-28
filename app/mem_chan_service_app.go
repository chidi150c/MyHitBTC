package app

import (
	"myhitbtcv4/model"
	"time"
)

type AppDB map[model.AppID]*App

func AppMemDBServiceFunc(AppMemDBChans MDDBChans, memAppDataChanChan chan chan *model.AppData) {
	var (
		dat AppDbServiceVehicle
		v   *App
		ok  bool
		mAppData AppVehicle
	)
	db := make(AppDB) //the mainDataBase very impotant
	every30mins := time.After(time.Minute * 2)
	for {
		select {
		case <-every30mins:			
			memAppDataChan := make(chan *model.AppData)
			memAppDataChanChan <- memAppDataChan
			for _, v = range db {
				mAppData = <-v.Chans.MyChan
				memAppDataChan <- mAppData.App.Data
				mAppData.RespChan <- true
			}
			close(memAppDataChan)
			every30mins = time.After(time.Minute * 30)
		case dat = <-AppMemDBChans.AddOrUpdateDbChan:
			db[dat.AppID] = dat.App
			if v, ok = db[dat.AppID]; ok && v.Data.ID == dat.AppID {
				dat.CallerChan <- AppDbResp{v.Data.ID, v, nil}
			} else {
				dat.CallerChan <- AppDbResp{Err: model.ErrInternal}
			}
		case dat = <-AppMemDBChans.GetDbChan:
			if v, ok = db[dat.AppID]; ok && v.Data.ID == dat.AppID {
				dat.CallerChan <- AppDbResp{v.Data.ID, v, nil}
			} else {
				dat.CallerChan <- AppDbResp{Err: model.ErrAppNotFound}
			}
		case dat = <-AppMemDBChans.DeleteDbChan:
			delete(db, dat.AppID)
			if v, ok = db[dat.AppID]; ok && v.Data.ID == dat.AppID {
				dat.CallerChan <- AppDbResp{Err: model.ErrInternal}
			} else {
				dat.CallerChan <- AppDbResp{}
			}
		}
	}
}

type AppMemDBService struct {
	session *Session
}

func (u *AppMemDBService) AddApp(App *App) error {
	CallerChan := make(chan AppDbResp)
	u.session.appMemDBChans.AddOrUpdateDbChan <- AppDbServiceVehicle{App.Data.ID, App, CallerChan}
	apDbResp := <-CallerChan
	return apDbResp.Err
}
func (u *AppMemDBService) GetApp(id model.AppID) (*App, error) {
	if id == 0 {
		return nil, model.ErrAppNameEmpty
	}
	CallerChan := make(chan AppDbResp)
	u.session.appMemDBChans.GetDbChan <- AppDbServiceVehicle{id, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.App != nil && apDbResp.Err == nil {
		return apDbResp.App, nil
	}
	return apDbResp.App, apDbResp.Err
}
func (u *AppMemDBService) UpdateApp(App *App) error {
	cachedUser, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update App.Data.
	CallerChan := make(chan AppDbResp)
	u.session.appMemDBChans.GetDbChan <- AppDbServiceVehicle{App.Data.ID, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.App != nil && apDbResp.Err == nil {
		if apDbResp.App.Data.ID != cachedUser.ApIDs[App.Data.SymbolCode] {
			return model.ErrUnauthorized
		}
		u.session.cachedApp = App
		u.session.appMemDBChans.AddOrUpdateDbChan <- AppDbServiceVehicle{App.Data.ID, App, CallerChan}
		apDbResp := <-CallerChan
		return apDbResp.Err
	}
	return model.ErrAppNotFound
}
func (u *AppMemDBService) DeleteApp(id model.AppID) error {
	user, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update App.Data.
	CallerChan := make(chan AppDbResp)
	u.session.appMemDBChans.GetDbChan <- AppDbServiceVehicle{id, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.App != nil && apDbResp.Err == nil {
		if apDbResp.App.Data.ID != user.ApIDs[apDbResp.App.Data.SymbolCode] {
			return model.ErrUnauthorized
		}
		u.session.cachedApp = &App{}
		u.session.appMemDBChans.DeleteDbChan <- AppDbServiceVehicle{id, nil, CallerChan}
		apDbResp := <-CallerChan
		return apDbResp.Err
	}
	return model.ErrAppNotFound
}

