# NymSocketManager

[![Go Reference](https://pkg.go.dev/badge/notrustverify/nymsocketmanager.svg)](https://pkg.go.dev/notrustverify/nymsocketmanager)

This Go module handles the connection to the Nym Mixnet let you focus on the rest of your application.

Note that this module needs a running nym-client to connect to the mixnet! It can be downloaded [here](https://nymtech.net/download-nym-components/) or built according the instructions [here](https://nymtech.net/docs/binaries/building-nym.html). Instructions for initiating and running a nym-client can be found [here](https://nymtech.net/docs/clients/websocket-client.html).

## Installation

NymSocketManager is available using the standard `go get` command.

Install it by running:
```bash
go get -u github.com/notrustverify/nymsocketmanager
```

## Usage

The module can be imported as following:

```go
import NymSocketManager "github.com/notrustverify/nymsocketmanager"
```

You can thenow instantiate the NymSocketManager or the SocketManager.

## Example

Examples on how to use both NymSocketManager and SocketManager can be found in the [examples](https://github.com/notrustverify/nymsocketmanager) folder.   
You can also check our Nostr-Nym proxy in Go: [NostrNym](https://github.com/notrustverify/nostr-nym).

## Future improvements

The following could be improved regarding this module:

* Improve type documentation
* Write more tests
* Use the [WS library](https://pkg.go.dev/github.com/gobwas/ws) for websocket connections. This module currently uses the [Gorilla Websocket library](https://pkg.go.dev/github.com/gorilla/websocket), which is unmaintained at the current time of writing (05.2023).

## License

This code is released under the GPLv3+ license.