package nymsocketmanager_test

import (
	"testing"

	lib "github.com/notrustverify/nymsocketmanager"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func emptyProcessing(lib.NymReceived, func(lib.NymMessage) error) {}

func TestNymSocketManagerShouldHaveANonEmptyConnectionURI(t *testing.T) {

	logger := zerolog.Logger{}

	_, e := lib.NewNymSocketManager("", emptyProcessing, &logger)
	require.Error(t, e)
}

func TestNymSocketManagerShouldHaveAValidProcessingFunction(t *testing.T) {

	logger := zerolog.Logger{}

	_, e := lib.NewNymSocketManager("ws://127.0.0.1", nil, &logger)
	require.Error(t, e)
}

func TestNymSocketManagerShouldHaveAValidLogger(t *testing.T) {
	_, e := lib.NewNymSocketManager("ws://127.0.0.1", emptyProcessing, nil)
	require.Error(t, e)
}

func TestNymSocketManagerShouldNotStartWithWrongClientIDWS(t *testing.T) {
	logger := zerolog.Logger{}

	nymSocketManager, e := lib.NewNymSocketManager("aaaaaaaaaaaa", emptyProcessing, &logger)
	require.NoError(t, e)

	_, e = nymSocketManager.Start()
	require.Error(t, e)
}

func TestNymSocketManagerGetGateway(t *testing.T) {

	logger := zerolog.Logger{}

	nymSocketManager, e := lib.NewNymSocketManager("ws://127.0.0.1:10977", func(_ lib.NymReceived, _ func(lib.NymMessage) error) {}, &logger)
	require.NoError(t, e)

	_, e = nymSocketManager.Start()
	require.NoError(t, e)
	nymSocketManager.Stop()

	gatewayAddr := nymSocketManager.GetConnectedGateway()
	require.NotEmpty(t, gatewayAddr)
}
