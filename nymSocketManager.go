package nymsocketmanager

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"golang.org/x/xerrors"
)

/*
 * This class is managing a listening socket which will hang on listening packets as well as manage a routine to send receiving packets to the Nym mixnet
 * The goal is to be more performant in case of high demand. Also, packets to the mixnet can come from both directions.
 */

func NewNymSocketManager(connectionURI string, messageHandler func(NymReceived, func(NymMessage) error), parentLogger *zerolog.Logger) (*NymSocketManager, error) {
	if len(connectionURI) == 0 {
		err := xerrors.Errorf("connection URI cannot be empty")
		return nil, err
	}

	if nil == messageHandler {
		err := xerrors.Errorf("processing function needs to be defined")
		return nil, err
	}

	if nil == parentLogger {
		err := xerrors.Errorf("logger needs to be defined")
		return nil, err
	}

	localLogger := parentLogger.With().Str(ComponentField, "NymSocketManager").Logger()

	return &NymSocketManager{
		connectionURI:  connectionURI,
		messageHandler: messageHandler,
		logger:         &localLogger,
	}, nil
}

type NymSocketManager struct {
	sync.Mutex

	clientID string

	connectionURI           string
	connection              *websocket.Conn
	selfInstanceStoppedChan chan struct{}

	// Related to listening
	socketListener           *SocketListener
	messageHandler           func(NymReceived, func(NymMessage) error)
	closedSocketListenerChan chan struct{}

	// Related to sender
	senderMutex sync.Mutex

	selfAddressReceivedChan chan interface{}

	logger *zerolog.Logger
}

func (n *NymSocketManager) IsRunning() bool {
	n.Lock()
	defer n.Unlock()
	return nil != n.connection
}

func (n *NymSocketManager) Start() (chan struct{}, error) {
	n.Lock()
	defer n.Unlock()

	n.logger.Debug().Msg("starting NymSocketManager")

	// Do not start if already started
	if nil != n.connection {
		n.logger.Warn().Msgf("connection to websocket %s already established. Resuming...", n.connectionURI)
		return nil, nil
	}

	// Open WS connection
	var e error
	n.connection, _, e = websocket.DefaultDialer.Dial(n.connectionURI, nil)
	if nil != e {
		err := xerrors.Errorf("failed to open connection to %v (%v). Is the websocket up and running?", n.connectionURI, e)
		n.logger.Warn().Msg(err.Error())
		return nil, err
	}

	// After which we start a listener for the packets
	n.socketListener, n.closedSocketListenerChan, e = NewSocketListener(n.connection, n.messageDispatcher, n.Stop, n.logger)
	if nil != e {
		err := xerrors.Errorf("failed to initiate the socketListener: %v", e)
		n.logger.Warn().Msg(err.Error())
		// Cancel progress so far
		n.selfDestruct()
		return nil, err
	}
	go n.socketListener.Listen()

	// To ensure everything works as expected, collect clientID

	// Create chan for messageDispatcher to indicate when response received
	n.selfAddressReceivedChan = make(chan interface{})

	e = n.Send(NewSelfAddressRequest())
	if nil != e {
		err := xerrors.Errorf("failed to send SelfAddressRequest: %v", e)
		n.logger.Warn().Msg(err.Error())

		// Cancel progress so far
		n.selfDestruct()
		return nil, err
	}

	timeout := time.After(5 * time.Second)
	select {
	case <-n.selfAddressReceivedChan:
		n.logger.Debug().Msgf("successfully collected clientID with socketListener")
		n.selfAddressReceivedChan = nil

	// Fail
	case <-timeout:
		err := xerrors.Errorf("failed to collect clientID from %v", n.connectionURI)
		n.logger.Warn().Msg(err.Error())
		// Cancel progress so far
		n.selfDestruct()
		return nil, err
	}

	n.selfInstanceStoppedChan = make(chan struct{}, 1)

	n.logger.Debug().Msg("started NymSocketManager")

	return n.selfInstanceStoppedChan, nil
}

func (n *NymSocketManager) Stop() {
	n.Lock()
	defer n.Unlock()

	n.logger.Debug().Msg("stopping NymSocketManager")

	// Check if not already fully stopped (setting connection to nil is last step of self-destruction)
	if nil == n.connection {
		return
	}

	n.selfDestruct()

	n.logger.Debug().Msg("stopped NymSocketManager")
}

// selfDestruct will close all channel and free resources when requested
// called from methods that already acquired the lock
func (n *NymSocketManager) selfDestruct() {

	n.logger.Debug().Msg("selfDestructing")

	// Ensure we do not close everthing if everything is closed already
	if nil == n.selfInstanceStoppedChan {
		n.logger.Debug().Msg("already selfDestructed")
		return
	}

	// How to properly close the connection (well, almost):
	///////////////////////////////////////////////////////
	/* This method properly close it from the other end's perspective
	 * on this side, it results in an abnormal closure, while we send a CloseNormalClosure message
	 * It seems to be an issue in this lib (ref: https://github.com/gorilla/websocket/pull/487).
	 */

	// If socketListener is defined, we close it
	if nil != n.socketListener {

		// This will close the socketListener
		n.logger.Trace().Msg("sending close signal on socket and waiting for confirmation from socketListener")
		n.sendCloseSignal()

		// Waiting for confirmation (or timeout)
		deadline := 5 * time.Second
		select {
		case <-n.closedSocketListenerChan:
			n.logger.Debug().Msg("underlying connection closed")
		case <-time.After(deadline):
			n.logger.Debug().Msgf("timed-out (%v) on waiting for underlying connection to close", deadline)
		}

		n.logger.Trace().Msg("removing socketListener")
		n.socketListener = nil
	}

	if nil != n.connection {
		n.logger.Trace().Msg("closing local connection")
		e := n.connection.Close()
		if e != nil {
			n.logger.Warn().Msgf("error while closing connection: %v", e)
		}
		n.connection = nil
	}

	// If initialized, we close the selfInstanceStoppedChan
	if nil != n.selfInstanceStoppedChan {
		n.logger.Trace().Msg("closing channel to indicate upstream that closed")
		close(n.selfInstanceStoppedChan)
		n.selfInstanceStoppedChan = nil
	}

	n.logger.Debug().Msg("selfDestructed")
}

// Send a message to the underlying connection
func (n *NymSocketManager) Send(msg NymMessage) error {
	n.senderMutex.Lock()
	defer n.senderMutex.Unlock()

	if nil == n.connection {
		err := xerrors.Errorf("connection is undefined. Is the NymSocketManager started?")
		n.logger.Warn().Msg(err.Error())
		return err
	}

	msgBytes, e := json.Marshal(msg)
	if nil != e {
		err := xerrors.Errorf("failed to marshal NymMessage: %v", msg)
		n.logger.Warn().Msg(err.Error())
		return err
	}

	e = n.connection.WriteMessage(websocket.TextMessage, msgBytes)
	if nil != e {
		err := xerrors.Errorf("failed to send message: %v", e)
		n.logger.Warn().Msg(err.Error())
		return err
	}

	return nil
}

// Send message to properly close the socket connection
// This will close any listener connected to this socket
func (n *NymSocketManager) sendCloseSignal() error {
	n.senderMutex.Lock()
	defer n.senderMutex.Unlock()

	if nil == n.connection {
		err := xerrors.Errorf("connection is undefined. Is the NymSocketManager started?")
		n.logger.Warn().Msg(err.Error())
		return err
	}

	e := n.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if nil != e {
		err := xerrors.Errorf("failed to write close: %v", e)
		n.logger.Warn().Msg(err.Error())
		return err
	}

	n.logger.Debug().Msg("sent websocket close message")

	return nil
}

func (n *NymSocketManager) GetNymClientId() string {
	n.Lock()
	defer n.Unlock()
	return n.clientID
}

// messageDispatcher is provided to the socketListener to process the incoming messages.
// It calls the provided messageHandler on received messages (except on errors and on selfAddress reply)
func (n *NymSocketManager) messageDispatcher(s []byte) {

	receivedMessageJSON := make(map[string]interface{})
	e := json.Unmarshal(s, &receivedMessageJSON)
	if nil != e {
		n.logger.Warn().Msgf("failed to unmarshal message: %v\n", e)
		return
	}

	if _, ok := receivedMessageJSON["type"]; !ok {
		n.logger.Warn().Msgf("message from mixnet have no \"type\" attribute. Message: %v", receivedMessageJSON)
		return
	}

	switch receivedMessageJSON["type"] {
	case NymSelfAddressReplyType:
		reply := NymSelfAddressReply{}
		e = json.Unmarshal(s, &reply)
		if nil != e {
			n.logger.Warn().Msgf("failed to unmarshal SelfAddressReply: %v", e)
			return
		}
		n.clientID = reply.Address
		n.logger.Debug().Msgf("Got %v reply: Address is %v", reply.Type, reply.Address)
		if nil != n.selfAddressReceivedChan {
			close(n.selfAddressReceivedChan)
		}

	case NymErrorType:
		reply := NymError{}
		e = json.Unmarshal(s, &reply)
		if nil != e {
			n.logger.Warn().Msgf("failed to unmarshal errorMessage: %v", e)
			return
		}
		n.logger.Error().Msgf("Got error from mixnet: %v", reply.Message)

	case NymReceivedType:
		msg := NymReceived{}
		e = json.Unmarshal(s, &msg)
		if nil != e {
			n.logger.Warn().Msgf("failed to unmarshal NymMessage: %v", e)
			return
		}
		n.logger.Debug().Msgf("got: %v", msg)

		n.messageHandler(msg, n.Send)

	default:
		n.logger.Warn().Msgf("encountered unparsed type of message: %v", receivedMessageJSON)
	}
}
