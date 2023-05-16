package nymsocketmanager_test

import (
	"math/rand"
	"testing"
	"time"

	lib "github.com/notrustverify/nymsocketmanager"
	"github.com/stretchr/testify/require"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

// According to https://stackoverflow.com/a/31832326
// Could be seriously improved
const charsForRandom = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charsForRandom[rand.Intn(len(charsForRandom))]
	}
	return string(b)
}

/*********************************************
 * NymError
 *********************************************/

func TestNymErrorNewEmtpy(t *testing.T) {
	n := lib.NymError{}.NewEmpty()

	require.Equal(t, n.(lib.NymError).Type, lib.NymErrorType)
}

/*********************************************
 * NymSelfAddressRequest
 *********************************************/

func TestNymSelfAddressRequestNewEmtpy(t *testing.T) {
	n := lib.NymSelfAddressRequest{}.NewEmpty()

	require.Equal(t, n.(lib.NymSelfAddressRequest).Type, lib.NymSelfAddressType)
}

/*********************************************
 * NymSelfAddressReply
 *********************************************/

func TestNymSelfAddressReplyNewEmtpy(t *testing.T) {
	n := lib.NymSelfAddressReply{}.NewEmpty()

	require.Equal(t, n.(lib.NymSelfAddressReply).Type, lib.NymSelfAddressType)
}

/*********************************************
 * NymMessage
 *********************************************/

func TestNymReceivedNewEmtpy(t *testing.T) {
	n := lib.NymReceived{}.NewEmpty()

	require.Equal(t, n.(lib.NymReceived).Type, lib.NymReceivedType)
}

/*********************************************
 * NymReply
 *********************************************/

func TestNymReplyNewEmtpy(t *testing.T) {
	n := lib.NymReply{}.NewEmpty()

	require.Equal(t, n.(lib.NymReply).Type, lib.NymReplyType)
}

func TestNewNymReplyCorrectlySetsValues(t *testing.T) {
	senderTag := RandStringBytes(5)
	message := RandStringBytes(5)

	n := lib.NewNymReply(senderTag, message)
	require.Equal(t, n.(lib.NymReply).SenderTag, senderTag)
	require.Equal(t, n.(lib.NymReply).Message, message)
}
