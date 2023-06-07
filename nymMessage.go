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

const NymSelfAddressType = "selfAddress"

func NewSelfAddressRequest() NymMessage {
	return NymSelfAddressRequest{
		NymMessageCommon{
			Type: NymSelfAddressType,
		},
	}
}

type NymSelfAddressRequest struct {
	NymMessageCommon
}

func (NymSelfAddressRequest) NewEmpty() NymMessage {
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

func NewSelfAddressReply(address string) NymMessage {
	return NymSelfAddressReply{
		NymMessageCommon{
			Type: NymSelfAddressReplyType,
		},
		address,
	}
}

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
 * NymSend
 *********************************************/

const NymSendType = "send"

func NewNymSend(message string, recipient string) NymMessage {
	return NymSend{
		NymMessageCommon{
			Type: NymSendType,
		},
		message, recipient,
	}
}

type NymSend struct {
	NymMessageCommon

	Message   string `json:"message"`
	Recipient string `json:"recipient"`
}

func (NymSend) NewEmpty() NymMessage {
	return NymSend{
		NymMessageCommon{
			Type: NymSendType,
		},
		"", "",
	}
}

func (NymSend) Name() string {
	return "NymSend"
}

func (n NymSend) String() string {
	s := fmt.Sprintf("NymSend to %s: %s", n.Recipient, n.Message)
	return s
}

/*********************************************
 * NymSendAnonymous
 *********************************************/

const NymSendAnonymousType = "sendAnonymous"

func NewNymSendAnonymous(message string, recipient string, nbReplySurbs uint) NymMessage {
	return NymSendAnonymous{
		NymMessageCommon{
			Type: NymSendType,
		},
		message, recipient, nbReplySurbs,
	}
}

type NymSendAnonymous struct {
	NymMessageCommon

	Message    string `json:"message"`
	Recipient  string `json:"recipient"`
	ReplySurbs uint   `json:"replySurbs"`
}

func (NymSendAnonymous) NewEmpty() NymMessage {
	return NymSendAnonymous{
		NymMessageCommon{
			Type: NymSendAnonymousType,
		},
		"", "", 0,
	}
}

func (NymSendAnonymous) Name() string {
	return "NymSendAnonymous"
}

func (n NymSendAnonymous) String() string {
	s := fmt.Sprintf("NymSendAnonymous to %s: %s with %d replySurbs", n.Recipient, n.Message, n.ReplySurbs)
	return s
}

/*********************************************
 * NymReceived
 *********************************************/

const NymReceivedType = "received"

func NewNymReceived(message string, senderTag string) NymMessage {
	return NymReceived{
		NymMessageCommon{
			Type: NymReceivedType,
		},
		message, senderTag,
	}
}

type NymReceived struct {
	NymMessageCommon

	Message   string `json:"message"`
	SenderTag string `json:"senderTag"`
}

func (NymReceived) NewEmpty() NymMessage {
	return NymReceived{
		NymMessageCommon{
			Type: NymReceivedType,
		},
		"", "",
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
		"", "",
	}
}

func (NymReply) Name() string {
	return NymReplyType
}

func (n NymReply) String() string {
	s := fmt.Sprintf("NymReply for %s: \"%s\"", n.SenderTag, n.Message)
	return s
}
