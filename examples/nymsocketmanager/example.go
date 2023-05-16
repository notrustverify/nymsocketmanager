package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"

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

	logger.Info().Msg("Starting the mixnet application")

	// Create NymSocketManager
	nymSocketManager, e := NymSocketManager.NewNymSocketManager(NYM_CLIENT_WS, msgHandler, &logger)
	if nil != e {
		logger.Error().Msgf("failed to create the NymSocketManager: %v", e)
		return
	}
	// Start the NymSocketManager and collect ClientID
	stoppedSocketManager, e := nymSocketManager.Start()
	if nil != e {
		logger.Error().Msgf("failed to start the NymSocketManager: %v", e)
		return
	}

	fmt.Printf("Client id is: %s\n", nymSocketManager.GetNymClientId())

	// Wait to be shut down or socket is closed (could be restarted if needed instead of closing the program)
	select {
	case <-stoppedSocketManager:
		stoppedSocketManager = nil
		logger.Debug().Msg("socketManager has stopped (underlying connection closed)")

	case <-interrupt:
		logger.Debug().Msg("received shutdown signal, closing...")
		nymSocketManager.Stop()
	}

	logger.Info().Msg("Leaving you now, enjoy your life!")
}

func msgHandler(msg NymSocketManager.NymReceived, sendToMixnet func(NymSocketManager.NymMessage) error) {
	fmt.Printf("Received from %v: \"%v\"\n", msg.SenderTag, msg.Message)

	if len(msg.SenderTag) != 0 {

		// Create a reply
		reply := NymSocketManager.NewNymReply(msg.SenderTag, msg.Message).(NymSocketManager.NymReply)

		e := sendToMixnet(reply)
		if nil != e {
			fmt.Printf("failed to send message to mixnet: %v", e)
		}

		fmt.Printf("Replied to %v: \"%v\"\n", reply.SenderTag, reply.Message)
	}
}
