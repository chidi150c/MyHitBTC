package bolt

import (
	"myhitbtcv4/bolt/internal"
	"encoding/binary"
	"log"
	"myhitbtcv4/model"
	"time"

	boltdb "github.com/boltdb/bolt"
)

func UserBoltDBServiceFunc(UserBoltDBChans model.UDBChans, uDBRCC chan chan *model.User) {
	log.Printf("userBoltServiceFunc started")
	var (
		dat  model.UserDbData
		datN model.UserDbByNameData
	)
	//Database initialization Starts...
	//Initializing bolt DB
	db, err := boltdb.Open("userDB.db", 0600, &boltdb.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	//This implements User bolt DB services
	userBoltServ := NewuserBoltServ(db)
	for {
		select {
		case uDBRC := <-uDBRCC:
			go userBoltServ.GetAllUser(uDBRC)
		case dat = <-UserBoltDBChans.AddDbChan:
			dat.UserID, err = userBoltServ.AddUser(dat.User)
			if err == nil {
				dat.CallerChan <- model.UserDbResp{dat.UserID, dat.User, nil}
			} else {
				log.Printf("UserBoltServiceFunc1 %v",err)
				dat.CallerChan <- model.UserDbResp{Err: err}
			}
		case dat = <-UserBoltDBChans.UpdateDbChan:
			err = userBoltServ.UpdateUser(dat.User)
			if err == nil {
				dat.CallerChan <- model.UserDbResp{dat.UserID, dat.User, nil}
			} else {
				log.Printf("UserBoltServiceFunc2 %v",err)
				dat.CallerChan <- model.UserDbResp{Err: err}
			}
		case dat = <-UserBoltDBChans.GetDbChan:
			dat.User, err = userBoltServ.GetUser(dat.UserID)
			if err == nil && dat.UserID == dat.User.ID {
				dat.CallerChan <- model.UserDbResp{dat.User.ID, dat.User, nil}
			} else {
				log.Printf("UserBoltServiceFunc3 %v",err)
				dat.CallerChan <- model.UserDbResp{Err: err}
			}
		case datN = <-UserBoltDBChans.GetDbByNameChan:
			datN.User, err = userBoltServ.GetUserByName(datN.Username)
			if err == nil && datN.Username == datN.User.Username {
				datN.CallerChan <- model.UserDbResp{datN.User.ID, datN.User, nil}
			} else {
				log.Printf("UserBoltServiceFunc4 %v",err)				
				datN.CallerChan <- model.UserDbResp{Err: err}
			}
		case dat = <-UserBoltDBChans.DeleteDbChan:
			err = userBoltServ.DeleteUser(dat.UserID)
			if err != nil {
				dat.CallerChan <- model.UserDbResp{Err: err}
			} else {
				log.Printf("UserBoltServiceFunc5 %v",err)
				dat.CallerChan <- model.UserDbResp{}
			}
		}
	}
}

type userBoltServ struct {
	db *boltdb.DB
}

func NewuserBoltServ(udb *boltdb.DB) *userBoltServ {
	a := &userBoltServ{
		db: udb,
	}
	// Start a writable transaction.
	tx, err := a.db.Begin(true)
	if err != nil {
		log.Fatalln("Error1 in bolt.Newsessio:38: ", err)
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucketIfNotExists([]byte("Users"))
	if err != nil {
		log.Fatalln("Error2 in bolt.Newsessio:45:", err)
	}

	// Commit the transaction and check for error.
	if err = tx.Commit(); err != nil {
		log.Fatalln("Error3 in bolt.Newsessio:50:", err)
	}
	return a
}
func (a *userBoltServ) UpdateUser(usr *model.User) error {
	musr := &model.User{}
	err := a.db.Update(func(tx *boltdb.Tx) error {
		// Retrieve the Users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("Users"))
		if v := b.Get(uitob(usr.ID)); v == nil {
			return model.ErrUserNotFound
		} else if err := internal.UnmarshalUser(v, musr); err != nil {
			return err
		}
		if musr.ID == usr.ID && musr.Username == usr.Username {
			// Marshal User data into bytes.
			buf, err := internal.MarshalUser(usr)
			if err != nil {
				return err
			}			
			// Persist bytes to Users bucket.
			return b.Put(uitob(usr.ID), buf)
		}	
		return model.ErrUnauthorized	
	})
	return err
}
func (a *userBoltServ) AddUser(usr *model.User) (model.UserID, error) {
	musr := &model.User{}
	err := a.db.Update(func(tx *boltdb.Tx) error {
		// Retrieve the Users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("Users"))
		if err := b.ForEach(func(k, v []byte) error {
			if err := internal.UnmarshalUser(v, musr); err == nil {
				if musr.ID == usr.ID || musr.Username == usr.Username {
					usr.ID = musr.ID
					return model.ErrUserExists
				}
			}
			return nil
		}); err != nil {
			return err
		}
		// Generate ID for the User.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := b.NextSequence()
		usr.ID = model.UserID(id)

		// Marshal User data into bytes.
		buf, err := internal.MarshalUser(usr)
		if err != nil {
			return err
		}

		// Persist bytes to Users bucket.
		return b.Put(uitob(usr.ID), buf)
	})
	if err != nil {
		return usr.ID, err
	}
	return usr.ID, nil
}
func (a *userBoltServ) GetUserByName(usrname string) (*model.User, error) {
	if usrname == "" {
		return nil, model.ErrUserNameRequired
	}
	var musr = &model.User{}
	err := a.db.View(func(tx *boltdb.Tx) error {
		b := tx.Bucket([]byte("Users"))
		if err := b.ForEach(func(k, v []byte) error {
			if err := internal.UnmarshalUser(v, musr); err == nil {
				if musr.Username == usrname {
					return model.ErrUserExists
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	if err == model.ErrUserExists {
		return musr, nil
	}
	return nil, model.ErrUserNotFound
}
func (a *userBoltServ) GetUser(id model.UserID) (*model.User, error) {
	if id == 0 {
		return nil, model.ErrUserRequired
	}
	var musr = &model.User{}
	err := a.db.View(func(tx *boltdb.Tx) error {
		v := tx.Bucket([]byte("Users")).Get(uitob(id))
		return internal.UnmarshalUser(v, musr)
	})
	if err != nil {
		return nil, err
	}
	if musr.ID == 0 {
		return nil, model.ErrUserNotFound
	}
	return musr, nil
}
func (a *userBoltServ) DeleteUser(id model.UserID) error {
	err := a.db.Update(func(tx *boltdb.Tx) error {
		// Retrieve the Users bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte("Users"))
		return b.Delete(uitob(id))
	})
	if err != nil {
		return err
	}
	return nil
}
func (a *userBoltServ) GetAllUser(uDBRC chan *model.User) {
	user := &model.User{}
	var err error
	i := 1
	for  {
		user, err = a.GetUser(model.UserID(i))
		if err != nil{
			close(uDBRC)
			return
		}
		i++
		uDBRC <- user
	}	
	return
}
// uitob returns an 8-byte big endian representation of v.
func uitob(v model.UserID) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
