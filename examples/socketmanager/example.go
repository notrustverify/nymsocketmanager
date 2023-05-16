package main

import (
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/xerrors"

	NymSocketManager "github.com/notrustverify/nymsocketmanager"
)

const NYM_CLIENT_WS = "ws://127.0.0.1:1977"

func main() {

	// Handle the shutdown of the signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Get logger
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.TraceLevel).
		With().Timestamp().Logger()

	logger.Info().Msg("Starting the WS application")

	// Create SocketManager
	socketManager, e := NymSocketManager.NewSocketManager(NYM_CLIENT_WS, func(msgBytes []byte, replyFunction func([]byte) error) {
		// Remove potential final '\n'
		msg := strings.TrimSuffix(string(msgBytes), "\n")

		logger.Info().Msgf("Received \"%v\"", msg)

		e := replyFunction([]byte(msg))
		if nil != e {
			err := xerrors.Errorf("error while replying: %v", e)
			logger.Err(err).Msg("")
		}

		logger.Info().Msgf("Replied: \"%v\"", msg)
	}, &logger)
	if nil != e {
		logger.Error().Msgf("failed to create the SocketManager: %v", e)
		return
	}

	// Start the NymSocketManager and collect ClientID
	stoppedSocketManager, e := socketManager.Start()
	if nil != e {
		logger.Error().Msgf("failed to start the SocketManager: %v", e)
		return
	}

	logger.Info().Msg("Connected!")

	// Wait to be shut down or socket is closed (could be restarted if needed instead of closing the program)
	select {
	case <-stoppedSocketManager:
		stoppedSocketManager = nil
		logger.Debug().Msg("socketManager has stopped (underlying connection closed)")

	case <-interrupt:
		logger.Debug().Msg("received shutdown signal, closing...")
		socketManager.Stop()
	}

	logger.Info().Msg("Leaving you now, enjoy your life!")
}
