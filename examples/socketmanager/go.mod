module example.com/example

go 1.20

replace github.com/notrustverify/nymsocketmanager => ../../Documents/Workspaces/NTV/NymSocketManager

require (
	github.com/notrustverify/nymsocketmanager v0.0.0-00010101000000-000000000000
	github.com/rs/zerolog v1.29.1
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20210927094055-39ccf1dd6fa6 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
)
