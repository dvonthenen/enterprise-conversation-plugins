module github.com/dvonthenen/enterprise-conversation-plugins/clients/statistical-client

go 1.18

require (
	github.com/dvonthenen/enterprise-reference-implementation v0.1.6
	github.com/dvonthenen/symbl-go-sdk v0.1.5
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f
	k8s.io/klog/v2 v2.90.0
)

require (
	github.com/dvonthenen/websocket v1.5.1-dyv.2 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gordonklaus/portaudio v0.0.0-20220320131553-cc649ad523c1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
)

replace github.com/r3labs/sse/v2 => github.com/dvonthenen/sse/v2 v2.0.0-20221222171132-1daa5f8b774c

// replace github.com/dvonthenen/symbl-go-sdk => ../../dvonthenen/symbl-go-sdk
// replace github.com/dvonthenen/websocket => ../../dvonthenen/websocket
// replace github.com/dvonthenen/websocketproxy => ../../dvonthenen/websocketproxy
// replace github.com/dvonthenen/sse => ../../r3labs/sse
