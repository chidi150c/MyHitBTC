package app

import (
	//"log"
	"myhitbtcv4/model"
)
type UserBoltDBService struct {
	session *Session
}
func (u *UserBoltDBService) AddUser(user *model.User) error {
	CallerChan := make(chan model.UserDbResp)
	u.session.userBoltDBChans.AddDbChan <- model.UserDbData{user.ID, user, CallerChan}
	usrDbResp := <-CallerChan
	return usrDbResp.Err
}
func (u *UserBoltDBService) GetUser(id model.UserID) (*model.User, error) {
	if id == 0 {
		return nil, model.ErrUserNameEmpty
	}
	CallerChan := make(chan model.UserDbResp)
	u.session.userBoltDBChans.GetDbChan <- model.UserDbData{id, nil, CallerChan}
	usrDbResp := <-CallerChan
	if usrDbResp.User != nil && usrDbResp.Err == nil {
		return usrDbResp.User, nil
	}
	return usrDbResp.User, usrDbResp.Err
}
func (u *UserBoltDBService) GetUserByName(usrname string) (*model.User, error) {
	if usrname == "0" {
		return nil, model.ErrUserNameEmpty
	}
	CallerChan := make(chan model.UserDbResp)
	u.session.userBoltDBChans.GetDbByNameChan <- model.UserDbByNameData{usrname, nil, CallerChan}
	usrDbResp := <-CallerChan
	if usrDbResp.User != nil && usrDbResp.Err == nil {
		return usrDbResp.User, nil
	}
	return usrDbResp.User, usrDbResp.Err
}
func (u *UserBoltDBService) UpdateUser(user *model.User) error {
	cachedUser, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update user.
	if user.ID != cachedUser.ID {
		return model.ErrUnauthorized
	}
	CallerChan := make(chan model.UserDbResp)
	u.session.userBoltDBChans.UpdateDbChan <- model.UserDbData{user.ID, user, CallerChan}
	usrDbResp := <-CallerChan
	return usrDbResp.Err
}
func (u *UserBoltDBService) DeleteUser(id model.UserID) error {
	cachedUser, err := u.session.Authenticate()
	if err != nil {
		return err
	}
	// Only allow owner to update user.
	CallerChan := make(chan model.UserDbResp)
	u.session.userBoltDBChans.GetDbChan <- model.UserDbData{id, nil, CallerChan}
	usrDbResp := <-CallerChan
	if usrDbResp.User != nil && usrDbResp.Err == nil {
		if usrDbResp.User.ID != cachedUser.ID {
			return model.ErrUnauthorized
		}
		u.session.userBoltDBChans.DeleteDbChan <- model.UserDbData{id, nil, CallerChan}
		usrDbResp := <-CallerChan
		return usrDbResp.Err
	}
	return model.ErrUserNotFound
}
