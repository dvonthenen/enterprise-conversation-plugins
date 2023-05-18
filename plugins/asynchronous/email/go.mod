module github.com/dvonthenen/enterprise-conversation-plugins/plugins/asynchronous/email

go 1.18

require (
	github.com/dvonthenen/enterprise-reference-implementation v0.1.9
	github.com/dvonthenen/symbl-go-sdk v0.1.8
	gopkg.in/mail.v2 v2.3.1
	k8s.io/klog/v2 v2.90.0
)

require (
	github.com/dvonthenen/rabbitmq-manager v0.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/rabbitmq/amqp091-go v1.5.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

replace github.com/r3labs/sse/v2 => github.com/dvonthenen/sse/v2 v2.0.0-20221222171132-1daa5f8b774c

// replace github.com/dvonthenen/enterprise-reference-implementation => ../../../../../dvonthenen/enterprise-reference-implementation
// replace github.com/dvonthenen/symbl-go-sdk => ../../../../../dvonthenen/symbl-go-sdk
// replace github.com/dvonthenen/websocket => ../../../../../dvonthenen/websocket
// replace github.com/dvonthenen/websocketproxy => ../../../../../dvonthenen/websocketproxy
// replace github.com/dvonthenen/sse => ../../r3labs/sse
