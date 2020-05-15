package app

import (
	"myhitbtcv4/model"
	"time"
)

type MDDBChans struct {
	AddOrUpdateDbChan chan AppDbServiceVehicle
	GetDbChan         chan AppDbServiceVehicle
	DeleteDbChan      chan AppDbServiceVehicle
}

type AppDB map[model.AppID]*App

func AppDBServiceFunc(AppDBChans MDDBChans, memAppDataChanChan chan chan *model.AppData) {
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
		case dat = <-AppDBChans.AddOrUpdateDbChan:
			db[dat.AppID] = dat.App
			if v, ok = db[dat.AppID]; ok && v.Data.ID == dat.AppID {
				dat.CallerChan <- AppDbResp{v.Data.ID, v, nil}
			} else {
				dat.CallerChan <- AppDbResp{Err: model.ErrInternal}
			}
		case dat = <-AppDBChans.GetDbChan:
			if v, ok = db[dat.AppID]; ok && v.Data.ID == dat.AppID {
				dat.CallerChan <- AppDbResp{v.Data.ID, v, nil}
			} else {
				dat.CallerChan <- AppDbResp{Err: model.ErrAppNotFound}
			}
		case dat = <-AppDBChans.DeleteDbChan:
			delete(db, dat.AppID)
			if v, ok = db[dat.AppID]; ok && v.Data.ID == dat.AppID {
				dat.CallerChan <- AppDbResp{Err: model.ErrInternal}
			} else {
				dat.CallerChan <- AppDbResp{}
			}
		}
	}
}

type AppDBService struct {
	session *Session
}

func (u *AppDBService) AddApp(App *App) error {
	CallerChan := make(chan AppDbResp)
	u.session.appDBChans.AddOrUpdateDbChan <- AppDbServiceVehicle{App.Data.ID, App, CallerChan}
	apDbResp := <-CallerChan
	return apDbResp.Err
}



func (u *AppDBService) GetApp(id model.AppID) (*App, error) {
	if id == 0 {
		return nil, model.ErrAppNameEmpty
	}
	CallerChan := make(chan AppDbResp)
	u.session.appDBChans.GetDbChan <- AppDbServiceVehicle{id, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.App != nil && apDbResp.Err == nil {
		return apDbResp.App, nil
	}
	return apDbResp.App, apDbResp.Err
}
func (u *AppDBService) UpdateApp(App *App) error {
	cachedUser, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update App.Data.
	CallerChan := make(chan AppDbResp)
	u.session.appDBChans.GetDbChan <- AppDbServiceVehicle{App.Data.ID, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.App != nil && apDbResp.Err == nil {
		if apDbResp.App.Data.ID != cachedUser.ApIDs[App.Data.SymbolCode] {
			return model.ErrUnauthorized
		}
		u.session.cachedApp = App
		u.session.appDBChans.AddOrUpdateDbChan <- AppDbServiceVehicle{App.Data.ID, App, CallerChan}
		apDbResp := <-CallerChan
		return apDbResp.Err
	}
	return model.ErrAppNotFound
}
func (u *AppDBService) DeleteApp(id model.AppID) error {
	user, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update App.Data.
	CallerChan := make(chan AppDbResp)
	u.session.appDBChans.GetDbChan <- AppDbServiceVehicle{id, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.App != nil && apDbResp.Err == nil {
		if apDbResp.App.Data.ID != user.ApIDs[apDbResp.App.Data.SymbolCode] {
			return model.ErrUnauthorized
		}
		u.session.cachedApp = &App{}
		u.session.appDBChans.DeleteDbChan <- AppDbServiceVehicle{id, nil, CallerChan}
		apDbResp := <-CallerChan
		return apDbResp.Err
	}
	return model.ErrAppNotFound
}

type AppDbServiceVehicle struct {
	AppID      model.AppID
	App        *App
	CallerChan chan AppDbResp
}

type AppDbResp struct {
	AppID model.AppID
	App   *App
	Err   error
}
