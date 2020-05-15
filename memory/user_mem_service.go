package memory

import (
	"myhitbtcv4/model"
)


func UserDBServiceFunc(UserDBChans model.UDBChans) {
	var (
		dat model.UserDbData
		datN model.UserDbByNameData
		v   *model.User
		ok  bool
	)
	db := make(map[model.UserID]*model.User) //the mainDataBase very impotant
	for {
		select {
		case dat = <-UserDBChans.AddDbChan:
			db[dat.UserID] = dat.User
			if v, ok = db[dat.UserID]; ok && v.ID == dat.UserID {
				dat.CallerChan <- model.UserDbResp{v.ID, v, nil}
			} else {
				dat.CallerChan <- model.UserDbResp{Err: model.ErrInternal}
			}
		case dat = <-UserDBChans.UpdateDbChan:
			db[dat.UserID] = dat.User
			if v, ok = db[dat.UserID]; ok && v.ID == dat.UserID {
				dat.CallerChan <- model.UserDbResp{v.ID, v, nil}
			} else {
				dat.CallerChan <- model.UserDbResp{Err: model.ErrInternal}
			}
		case dat = <-UserDBChans.GetDbChan:
			if v, ok = db[dat.UserID]; ok && v.ID == dat.UserID {
				dat.CallerChan <- model.UserDbResp{v.ID, v, nil}
			} else {
				dat.CallerChan <- model.UserDbResp{Err: model.ErrInternal}
			}
		case datN = <-UserDBChans.GetDbByNameChan:
			b := true
			for _, v := range db{
				if v.Username == datN.Username{
					datN.CallerChan <- model.UserDbResp{v.ID, v, nil}
					b = false
					break
				} 
			}
			if b {
				datN.CallerChan <- model.UserDbResp{Err: model.ErrInternal}
			}
		case dat = <-UserDBChans.DeleteDbChan:
			delete(db, dat.UserID)
			if v, ok = db[dat.UserID]; ok && v.ID == dat.UserID {
				dat.CallerChan <- model.UserDbResp{Err: model.ErrInternal}
			} else {
				dat.CallerChan <- model.UserDbResp{}
			}
		}
	}
}