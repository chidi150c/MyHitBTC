package apichart

import (
	"time"
	"user-apiv2/apiuser"
	"user-apiv2/model"
)

/*Session has the database handle the services can reference them. By making the
chartService a non-pointer field we reduce the allocations required when creating
a new session.*/
type Session struct {
	//db interface{}
	session      model.Sessioner
	Graphservice GraphService
	//GraphPointChan  chan GraphPoint
	CGraphPointChan chan CGraphPoint
	//NumOfGraphPoints int
	GraphPointTime   time.Duration
	StringGraphChan  chan string
	CStringGraphChan chan string
	ReadFromFileChan chan string
	//Sendmore        chan bool
}

//uDB GDBType, us *apiuser.Session, gg *chartGuiService
func NewSession(dbUser apiuser.UDBType, msess model.Sessioner) *Session {
	s := &Session{
		//db:      aDB,
		session: msess,
		//GraphPointChan:  make(chan GraphPoint, 23),
		CGraphPointChan: make(chan CGraphPoint, 2),
		//NumOfGraphPoints: 24,
		GraphPointTime: time.Minute * 2,
		// Sendmore:        make(chan bool), //
		StringGraphChan:  make(chan string, 1),
		CStringGraphChan: make(chan string, 2),
		ReadFromFileChan: make(chan string, 1),
	}
	s.Graphservice.session = s
	return s
}
