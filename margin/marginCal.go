package margin

import (
	//"log"
	"myhitbtcv4/model"
)

func MarginCal(mCDBChan model.MCalDBChans) {
	var dat model.MarginDBVeh
	var appID model.AppID
	var MarginDB map[model.AppID]model.MarginParam
	marginRegister := make(map[model.AppID]chan model.MarginParam)
	for {
		MarginDB =  make(map[model.AppID]model.MarginParam)
		select {
		case dat = <-mCDBChan.AddMarginDBChan:
			marginRegister[dat.ID] = dat.MPChan
			dat.MCalDBRespChan <-model.MarginDBResp{Err: nil}
		case dat = <-mCDBChan.DeleteMarginDBChan:
			if _, ok := marginRegister[dat.ID]; ok{
				delete(marginRegister, dat.ID)
				dat.MCalDBRespChan <-model.MarginDBResp{Err: nil}
			}else{
				dat.MCalDBRespChan <-model.MarginDBResp{Err: model.ErrAppNotFound}
			}			
		case dat = <-mCDBChan.GetMarginDBChan:
			for _, appID = range dat.User.ApIDs {
				if MParamChan, ok := marginRegister[appID]; ok != false{
					MarginDB[appID] =  <-MParamChan
				}
			}
			if len(MarginDB) != 0 {
				dat.MCalDBRespChan <- model.MarginDBResp{
					ID: dat.User.ID,
					MarginDB: MarginDB, 
					Err: nil,
				}
			} else{
				dat.MCalDBRespChan <- model.MarginDBResp{
					ID: dat.User.ID,
					MarginDB: MarginDB, 
					Err: model.ErrAppNotFound,
				}
			}
		}
	}
}
