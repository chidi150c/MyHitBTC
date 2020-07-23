package app

import (
	"myhitbtcv4/model"
)

type AppBoltDBService struct {
	session *Session
}


func (u *AppBoltDBService) AddApp(md *model.AppData) (model.AppID, error) {
	CallerChan := make(chan model.AppDataResp)
	u.session.appBoltDBChans.AddDbChan <- model.AppDataBoltVehicle{md.ID, md, CallerChan}
	apDbResp := <-CallerChan
	return apDbResp.AppID, apDbResp.Err
}
func (u *AppBoltDBService) GetApp(id model.AppID) (*model.AppData, error) {
	if id == 0 {
		return nil, model.ErrAppNameEmpty
	}
	CallerChan := make(chan model.AppDataResp)
	u.session.appBoltDBChans.GetDbChan <- model.AppDataBoltVehicle{id, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.AppData != nil && apDbResp.Err == nil {
		return apDbResp.AppData, nil
	}
	return apDbResp.AppData, apDbResp.Err
}
func (u *AppBoltDBService) UpdateApp(md *model.AppData) error {
	cachedUser, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update md.
	CallerChan := make(chan model.AppDataResp)
	u.session.appBoltDBChans.GetDbChan <- model.AppDataBoltVehicle{md.ID, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.AppData != nil && apDbResp.Err == nil {
		if apDbResp.AppData.ID != cachedUser.ApIDs[md.SymbolCode] {
			return model.ErrUnauthorized
		}
		u.session.appBoltDBChans.UpdateDbChan <- model.AppDataBoltVehicle{md.ID, md, CallerChan}
		apDbResp := <-CallerChan
		return apDbResp.Err
	}
	return model.ErrAppNotFound
}
func (u *AppBoltDBService) DeleteApp(id model.AppID) error {
	user, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update md.
	CallerChan := make(chan model.AppDataResp)
	u.session.appBoltDBChans.GetDbChan <- model.AppDataBoltVehicle{id, nil, CallerChan}
	apDbResp := <-CallerChan
	if apDbResp.AppData != nil && apDbResp.Err == nil {
		if apDbResp.AppData.ID != user.ApIDs[apDbResp.AppData.SymbolCode] {
			return model.ErrUnauthorized
		}
		u.session.appBoltDBChans.DeleteDbChan <- model.AppDataBoltVehicle{id, nil, CallerChan}
		apDbResp := <-CallerChan
		return apDbResp.Err
	}
	return model.ErrAppNotFound
}

