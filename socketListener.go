package nymsocketmanager

import (
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"golang.org/x/xerrors"
)

func NewSocketListener(socket *websocket.Conn, messageHandler func([]byte), toCallWhenClosed func(), parentLogger *zerolog.Logger) (*SocketListener, chan struct{}, error) {

	if nil == socket {
		err := xerrors.Errorf("websocket connection cannot be undefined")
		return nil, nil, err
	}

	if nil == messageHandler {
		err := xerrors.Errorf("processing function needs to be defined")
		return nil, nil, err
	}

	// toCallWhenClosed function can be nil if nothing needs to be done

	if nil == parentLogger {
		err := xerrors.Errorf("logger needs to be defined")
		return nil, nil, err
	}

	closedSocketChan := make(chan struct{}, 1)

	localLogger := parentLogger.With().Str(ComponentField, "SocketListener").Logger()

	return &SocketListener{
		socket:           socket,
		closedSocketChan: closedSocketChan,
		logger:           &localLogger,
		messageHandler:   messageHandler,
		toCallWhenClosed: toCallWhenClosed,
	}, closedSocketChan, nil
}

type SocketListener struct {
	socket *websocket.Conn

	messageHandler func([]byte)

	toCallWhenClosed func()

	closedSocketChan chan struct{}
	logger           *zerolog.Logger
}

func (s *SocketListener) Listen() {

	// If provided, execute some cleaning code from parent after closing
	if s.toCallWhenClosed != nil {
		s.logger.Trace().Msg("socketListener instructed to call function when shutting down")
		defer s.toCallWhenClosed()
	}

	for nil != s.socket {
		_, receivedMessage, e := s.socket.ReadMessage()
		if nil != e {
			s.logger.Debug().Msgf("Read: \"%v\"", e)
			break
		}

		// Process msg: start a goroutine to handle the request
		s.logger.Trace().Msgf("recv: \"%s\"", string(receivedMessage))
		go s.messageHandler(receivedMessage)
	}

	// When the connection will be closed, will close the chan
	// so that main process can know that ws has been closed.
	close(s.closedSocketChan)
	s.closedSocketChan = nil

	s.logger.Debug().Msg("socketListener shut down")
}
