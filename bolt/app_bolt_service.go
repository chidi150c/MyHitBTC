package bolt

import (
	"encoding/binary"
	"myhitbtcv4/bolt/internal"
	"log"
	"myhitbtcv4/model"
	"time"

	boltdb "github.com/coreos/bbolt"
)

func AppDataBoltDBServiceFunc(AppDataBoltDBChans model.ABDBChans, memAppDataChanChan chan chan *model.AppData) {
	log.Printf("AppDataBoltServiceFunc started")
	//Database initialization Starts...
	//Initializing bolt DB
	db, err := boltdb.Open("data.db", 0600, &boltdb.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var (
		dat     model.AppDataBoltVehicle
		id      model.AppID
		appData *model.AppData
	)
	//This implements AppData bolt DB services
	appBoltServ := NewAppDataService(db)
	for {
		select {
		case dat = <-AppDataBoltDBChans.AddDbChan:
			id, err = appBoltServ.AddAppData(dat.AppData)
			if err == nil && id != 0 {
				dat.AppData.ID = id
				dat.CallerChan <- model.AppDataResp{id, dat.AppData, nil}
			} else {
				dat.CallerChan <- model.AppDataResp{id, nil, err}
			}
		case dat = <-AppDataBoltDBChans.UpdateDbChan:
			err = appBoltServ.UpdateAppData(dat.AppData)
			if err == nil {
				dat.CallerChan <- model.AppDataResp{dat.AppData.ID, dat.AppData, nil}
				log.Printf("DB UpdateAppData %v done id: %v", dat.AppData.SymbolCode, dat.AppData.ID)
			} else {
				dat.CallerChan <- model.AppDataResp{Err: model.ErrInternal}
			}
		case dat = <-AppDataBoltDBChans.GetDbChan:
			dat.AppData, err = appBoltServ.GetAppData(dat.AppID)
			if err == nil {
				dat.CallerChan <- model.AppDataResp{dat.AppData.ID, dat.AppData, nil}
			} else {
				dat.CallerChan <- model.AppDataResp{Err: model.ErrInternal}
			}
		case dat = <-AppDataBoltDBChans.DeleteDbChan:
			err = appBoltServ.DeleteAppData(dat.AppID)
			if err == nil {
				dat.CallerChan <- model.AppDataResp{dat.AppID, nil, nil}
			} else {
				dat.CallerChan <- model.AppDataResp{Err: model.ErrInternal}
			}
		case memoryAppDataChan := <-memAppDataChanChan:
			for appData = range memoryAppDataChan {
				err = appBoltServ.UpdateAppData(appData)
				if err != nil {
					log.Printf("DB Update Error: %v", err)
				}
			}
		}
	}
}

type AppDataService struct {
	db *boltdb.DB
}

func NewAppDataService(udb *boltdb.DB) *AppDataService {
	a := &AppDataService{
		db: udb,
	}
	// Start a writable transaction.
	tx, err := a.db.Begin(true)
	if err != nil {
		log.Fatalln("Error1 in bolt.Newsessio:38: ", err)
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucketIfNotExists([]byte("AppDatas"))
	if err != nil {
		log.Fatalln("Error2 in bolt.Newsessio:45:", err)
	}

	// Commit the transaction and check for error.
	if err = tx.Commit(); err != nil {
		log.Fatalln("Error3 in bolt.Newsessio:50:", err)
	}
	return a
}
func (a *AppDataService) UpdateAppData(appdat *model.AppData) error {
	err := a.db.Update(func(tx *boltdb.Tx) error {
		// Retrieve the AppDatas bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("AppDatas"))
		// Marshal AppData data into bytes.
		buf, err := internal.MarshalAppData(appdat)
		if err != nil {
			return err
		}
		// Persist bytes to AppDatas bucket.
		return b.Put(itob(appdat.ID), buf)
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *AppDataService) AddAppData(appdat *model.AppData) (model.AppID, error) {
	mappdat := &model.AppData{}
	err := a.db.Update(func(tx *boltdb.Tx) error {
		// Retrieve the AppDatas bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("AppDatas"))
		if err := b.ForEach(func(k, v []byte) error {
			if err := internal.UnmarshalAppData(v, mappdat); err == nil {
				if (mappdat.ID == appdat.ID) && (mappdat.SymbolCode == appdat.SymbolCode) && (mappdat.UsrID == appdat.UsrID) {
					appdat.ID = mappdat.ID
					return model.ErrAppExists
				}
			}
			return nil
		}); err != nil {
			return err
		}
		// Generate ID for the AppData.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := b.NextSequence()
		appdat.ID = model.AppID(id)

		// Marshal AppData data into bytes.
		buf, err := internal.MarshalAppData(appdat)
		if err != nil {
			return err
		}

		// Persist bytes to AppDatas bucket.
		return b.Put(itob(appdat.ID), buf)
	})
	if err != nil {
		return appdat.ID, err
	}
	return appdat.ID, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v model.AppID) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func (a *AppDataService) GetAppData(id model.AppID) (*model.AppData, error) {
	if id == 0 {
		return nil, model.ErrAppRequired
	}
	var mappdat = &model.AppData{}
	err := a.db.View(func(tx *boltdb.Tx) error {
		v := tx.Bucket([]byte("AppDatas")).Get(itob(id))
		return internal.UnmarshalAppData(v, mappdat)
	})
	if err != nil {
		return nil, err
	}
	if mappdat.ID == 0 {
		return nil, model.ErrAppNotFound
	}
	return mappdat, nil
}

func (a *AppDataService) DeleteAppData(id model.AppID) error {
	err := a.db.Update(func(tx *boltdb.Tx) error {
		// Retrieve the AppDatas bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("AppDatas"))
		return b.Delete(itob(id))
	})
	if err != nil {
		return err
	}
	return nil
}
