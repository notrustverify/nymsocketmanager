package nymsocketmanager

import "fmt"

type NymMessageCommon struct {
	Type string `json:"type"`
}

// Message defines the type of message that can be exchanged
type NymMessage interface {
	NewEmpty() NymMessage
	Name() string // Type of message
	String() string
}

/*********************************************
 * NymError
 *********************************************/

const NymErrorType = "error"

type NymError struct {
	NymMessageCommon

	Message string `json:"message"`
}

func (NymError) NewEmpty() NymMessage {
	return NymError{
		NymMessageCommon{
			Type: "error",
		},
		"",
	}
}

func (NymError) Name() string {
	return "error"
}

func (n NymError) String() string {
	s := fmt.Sprintf("NymError: %v", n.Message)
	return s
}

/*********************************************
 * NymSelfAddressRequest
 *********************************************/

func NewSelfAddressRequest() NymMessage {
	return NymSelfAddressRequest{
		NymMessageCommon{
			Type: NymSelfAddressType,
		},
	}
}

const NymSelfAddressType = "selfAddress"

type NymSelfAddressRequest struct {
	NymMessageCommon
}

func (s NymSelfAddressRequest) NewEmpty() NymMessage {
	return NewSelfAddressRequest()
}

func (NymSelfAddressRequest) Name() string {
	return "NymSelfAddressRequest"
}

func (NymSelfAddressRequest) String() string {
	return "NymSelfAddressRequest"
}

/*********************************************
 * NymSelfAddressReply
 *********************************************/

const NymSelfAddressReplyType = "selfAddress"

type NymSelfAddressReply struct {
	NymMessageCommon

	Address string `json:"address"`
}

func (s NymSelfAddressReply) NewEmpty() NymMessage {
	return NymSelfAddressReply{
		NymMessageCommon{
			Type: NymSelfAddressReplyType,
		},
		"",
	}
}

func (NymSelfAddressReply) Name() string {
	return "NymSelfAddressReply"
}

func (n NymSelfAddressReply) String() string {
	s := fmt.Sprintf("NymSelfAddressReply: %v", n.Address)
	return s
}

/*********************************************
 * NymMessage
 *********************************************/

const NymReceivedType = "received"

type NymReceived struct {
	NymMessageCommon

	Message   string `json:"message"`
	SenderTag string `json:"senderTag"`
}

func (n NymReceived) NewEmpty() NymMessage {
	return NymReceived{
		NymMessageCommon{
			Type: NymReceivedType,
		},
		"",
		"",
	}
}

func (NymReceived) Name() string {
	return NymReceivedType
}

func (n NymReceived) String() string {
	s := fmt.Sprintf("NymReceivedMessage from %v: \"%v\"", n.SenderTag, n.Message)
	return s
}

/*********************************************
 * NymReply
 *********************************************/

func NewNymReply(senderTag string, message string) NymMessage {
	return NymReply{
		NymMessageCommon{
			Type: NymReplyType,
		},
		message, senderTag,
	}
}

const NymReplyType = "reply"

type NymReply struct {
	NymMessageCommon

	Message   string `json:"message"`
	SenderTag string `json:"senderTag"`
}

func (n NymReply) NewEmpty() NymMessage {
	return NymReply{
		NymMessageCommon{
			Type: NymReplyType,
		},
		"",
		"",
	}
}

func (NymReply) Name() string {
	return NymReplyType
}

func (n NymReply) String() string {
	s := fmt.Sprintf("NymReply for %v: \"%v\"", n.SenderTag, n.Message)
	return s
}
