package nymsocketmanager

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"golang.org/x/xerrors"
)

func NewSocketManager(connectionURI string, messageHandler func([]byte, func([]byte) error), parentLogger *zerolog.Logger) (*SocketManager, error) {
	if len(connectionURI) == 0 {
		err := xerrors.Errorf("connection URI cannot be empty")
		return nil, err
	}

	if nil == parentLogger {
		err := xerrors.Errorf("logger needs to be defined")
		return nil, err
	}

	socketLogger := parentLogger.With().Str(ComponentField, "SocketManager").Logger()

	return &SocketManager{
		connectionURI:  connectionURI,
		messageHandler: messageHandler,
		logger:         &socketLogger,
	}, nil
}

type SocketManager struct {
	sync.Mutex

	connectionURI           string
	connection              *websocket.Conn
	selfInstanceStoppedChan chan struct{}

	// Related to listening
	socketListener           *SocketListener
	messageHandler           func([]byte, func([]byte) error)
	closedSocketListenerChan chan struct{}

	// Related to sending
	senderMutex sync.Mutex

	logger *zerolog.Logger
}

func (s *SocketManager) IsRunning() bool {
	s.Lock()
	defer s.Unlock()
	return nil != s.connection
}

func (s *SocketManager) Start() (chan struct{}, error) {
	s.Lock()
	defer s.Unlock()

	s.logger.Debug().Msg("starting SocketManager")

	// Do not start if already started
	if nil != s.connection {
		s.logger.Warn().Msgf("connection to websocket %s already established. Resuming...", s.connectionURI)
		return nil, nil
	}

	// Open WS connection
	var e error
	s.connection, _, e = websocket.DefaultDialer.Dial(s.connectionURI, nil)
	if nil != e {
		err := xerrors.Errorf("failed to open connection to \"%v\". Is the websocket up and running?", s.connectionURI)
		s.logger.Warn().Msg(err.Error())
		return nil, err
	}
	s.logger.Debug().Msgf("successfully opened connection to \"%v\"", s.connectionURI)

	// After which we start a listener for the packets
	s.socketListener, s.closedSocketListenerChan, e = NewSocketListener(s.connection, func(msg []byte) {
		s.messageHandler(msg, s.Send)
	}, s.Stop, s.logger)
	if nil != e {
		err := xerrors.Errorf("failed to initiate the socketListener: %v", e)
		s.logger.Warn().Msg(err.Error())
		// Cancel progress so far
		s.selfDestruct()
		return nil, err
	}
	go s.socketListener.Listen()

	s.selfInstanceStoppedChan = make(chan struct{}, 1)

	s.logger.Debug().Msg("started SocketManager")

	return s.selfInstanceStoppedChan, nil
}

func (s *SocketManager) Stop() {
	s.Lock()
	defer s.Unlock()

	s.logger.Debug().Msg("stopping SocketManager")

	// Do not stop if not running
	if nil == s.connection {
		return
	}

	s.selfDestruct()

	s.logger.Debug().Msg("stopped SocketManager")
}

// selfDestruct will close all channel and free resources when requested
// called from methods that already acquired the lock
func (s *SocketManager) selfDestruct() {

	s.logger.Debug().Msg("selfDestructing")

	// Ensure we do not close everthing if everything is closed already
	if nil == s.selfInstanceStoppedChan {
		s.logger.Warn().Msg("already selfDestructed (selfInstanceStoppedChan is nil)")
		return
	}

	// How to properly close the connection (well, almost):
	///////////////////////////////////////////////////////
	/* This method properly close it from the other end's perspective
	 * on this side, it results in an abnormal closure, while we send a CloseNormalClosure message
	 * It seems to be an issue in this lib (ref: https://github.com/gorilla/websocket/pull/487).
	 */

	// If socketListener is defined, we close it
	if nil != s.socketListener {

		// This will close the socketListener
		s.logger.Trace().Msg("sending close signal on socket and waiting for confirmation")
		s.sendCloseSignal()

		// Waiting for confirmation (or timeout)
		deadline := 5 * time.Second
		select {
		case <-s.closedSocketListenerChan:
			s.logger.Debug().Msg("underlying connection closed")
		case <-time.After(deadline):
			s.logger.Debug().Msgf("timed-out (%v) on waiting for underlying connection to close", deadline)
		}

		s.logger.Trace().Msg("removing socketListener")
		s.socketListener = nil
	}

	if nil != s.connection {
		s.logger.Trace().Msg("closing local connection")
		e := s.connection.Close()
		if e != nil {
			s.logger.Err(e).Msg("")
		}
		s.connection = nil
	}

	// If initialized, we close the selfInstanceStoppedChan
	if nil != s.selfInstanceStoppedChan {
		s.logger.Trace().Msg("closing channel to indicate upstream that closed")
		close(s.selfInstanceStoppedChan)
		s.selfInstanceStoppedChan = nil
	}

	s.logger.Debug().Msg("selfDestructed")
}

func (s *SocketManager) Send(message []byte) error {
	s.senderMutex.Lock()
	defer s.senderMutex.Unlock()

	if nil == s.connection {
		err := xerrors.Errorf("connection is closed, cannot send to %v", s.connectionURI)
		s.logger.Warn().Msg(err.Error())
		return err
	}

	e := s.connection.WriteMessage(websocket.TextMessage, message)
	if nil != e {
		err := xerrors.Errorf("failed to send message: %v", e)
		s.logger.Warn().Msg(err.Error())
		return err
	}

	return nil
}

// Send message to properly close the socket connection
// This will close any listener connected to this socket
func (s *SocketManager) sendCloseSignal() error {
	s.senderMutex.Lock()
	defer s.senderMutex.Unlock()

	if nil == s.connection {
		err := xerrors.Errorf("connection is undefined. Is the SocketManager started?")
		s.logger.Warn().Msg(err.Error())
		return err
	}

	e := s.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if nil != e {
		err := xerrors.Errorf("failed to write close: %v", e)
		s.logger.Warn().Msg(err.Error())
		return err
	}

	s.logger.Debug().Msg("sent websocket close message")

	return nil
}
