package socketapi

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"

	"github.com/thakkarparth007/dalal-street-server/session"
	"github.com/thakkarparth007/dalal-street-server/socketapi/actions"
	"github.com/thakkarparth007/dalal-street-server/utils"
)

var socketApiLogger *logrus.Entry
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return utils.Configuration.Stage == "test" || utils.Configuration.Stage == "dev"
	},
}

func InitSocketApi() {
	socketApiLogger = utils.Logger.WithFields(logrus.Fields{
		"module": "socketapi/SocketHandler",
	})
	actions.InitActions()
}

func loadSession(r *http.Request) (session.Session, error) {
	var l = socketApiLogger.WithFields(logrus.Fields{
		"method": "loadSession",
	})

	sidCookie, _ := r.Cookie("sid")
	if sidCookie != nil {
		l.Debugf("Found sid cookie")
		s, err := session.Load(sidCookie.Value)

		if err != nil {
			l.Errorf("Error loading session data: '%s'", err)
			return nil, err
		} else {
			l.Debugf("Loaded session")
			return s, nil
		}
	}

	s, err := session.New()
	if err != nil {
		l.Errorf("Error starting new session: '%s'", err)
		return nil, err
	}
	l.Debugf("Created new session")

	return s, nil
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var l = socketApiLogger.WithFields(logrus.Fields{
		"method": "Handle",
	})

	l.Infof("Connection from %+v", r.RemoteAddr)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.Errorf("Could not upgrade connection: '%s'", err)
		return
	}
	l.Debugf("Upgraded to websocket protocol")

	sess, err := loadSession(r)
	if err != nil {
		l.Errorf("Could not load or create session. Replying with 500. '%s'", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := NewClient(make(chan struct{}), make(chan []byte, 256), conn, sess)

	go c.WritePump()
	c.ReadPump()
}
