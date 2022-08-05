package posixmq

import (
	"io"
	"unsafe"

	"golang.org/x/sys/unix"
)

const MessageQueueDefaultMode int = 0644

type MessageQueue struct {
	io.Closer
	fd         int
	Name       string // Name of the message queue
	Attributes MessageQueueAttributes
}

type MessageQueueAttributes struct {
	Flags           int64 // Message queue flags
	MaxQueueSize    int64 // Max # of messages in queue
	MaxMessageSize  int64 // Max message size in bytes
	CurrentMessages int64 // Current # of messages in queue
}

func Open(name string, flags int64, cfg MessageQueueAttributes) (*MessageQueue, error) {
	unixName, err := unix.BytePtrFromString(name)
	if err != nil {
		return nil, err
	}

	mqfd, _, errno := unix.Syscall6(
		unix.SYS_MQ_OPEN,
		uintptr(unsafe.Pointer(unixName)),
		uintptr(flags),
		uintptr(MessageQueueDefaultMode),
		uintptr(unsafe.Pointer(&cfg)),
		0, // Last 2 unused
		0,
	)

	if errno != 0 {
		return nil, errno
	}

	return &MessageQueue{
		fd:         int(mqfd),
		Name:       name,
		Attributes: cfg,
	}, nil
}
