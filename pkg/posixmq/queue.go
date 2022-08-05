package posixmq

import "io"

const MessageQueueDefaultMode int = 0644

type MessageQueue struct {
	io.Closer
	fd         uintptr
	Name       string // Name of the message queue
	Attributes MessageQueueAttributes
}

type MessageQueueAttributes struct {
	Flags           int64 // Message queue flags
	MaxQueueSize    int64 // Max # of messages in queue
	MaxMessageSize  int64 // Max message size in bytes
	CurrentMessages int64 // Current # of messages in queue
}
