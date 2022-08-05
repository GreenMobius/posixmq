package posixmq

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

const MessageQueueDefaultMode int = 0644
const MessageQueueMaxQueueSize int64 = 10
const MessageQueueMaxMessageSize int64 = 8192

type MessageQueue struct {
	fd         int
	name       string // Name of the message queue
	attributes MessageQueueAttributes
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
		name:       name,
		attributes: cfg,
	}, nil
}

func (mq *MessageQueue) Close() error {
	return unix.Close(int(mq.fd))
}

func (mq *MessageQueue) Unlink() error {
	unixName, err := unix.BytePtrFromString(mq.name)
	if err != nil {
		return err
	}

	_, _, errno := unix.Syscall(
		unix.SYS_MQ_UNLINK,
		uintptr(unsafe.Pointer(unixName)),
		0, // Last 2 unused
		0,
	)
	if errno != 0 {
		return errno
	}

	return nil
}
