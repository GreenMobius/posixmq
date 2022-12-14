package posixmq

import (
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const MessageQueueDefaultMode int = 0644
const MessageQueueMaxQueueSize int64 = 10
const MessageQueueMaxMessageSize int64 = 8192

// ErrMessageQueueFull indicates that a send failed because a message queue is full
var ErrMessageQueueFull = errors.New("message queue is full")

// ErrMessageQueueEmpty indicates that a receive failed because a message queue is empty
var ErrMessageQueueEmpty = errors.New("message queue is empty")

// ErrMessageTooLarge indicates that a send failed because a message was longer than the specified maximum size
var ErrMessageTooLarge = errors.New("message exceeds maximum size")

// ErrMessageQueueInvalid indicates that a message queue is not valid
var ErrMessageQueueInvalid = errors.New("invalid message queue")

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

func (mq *MessageQueue) commonSend(msg []byte, priority uint, timeout *time.Duration) error {
	if len(msg) == 0 {
		return errors.New("sending empty messages is not supported")
	}

	var timeoutPtr uintptr = 0
	if timeout != nil {
		deadline := time.Now().Add(*timeout)
		unixTimeout, err := unix.TimeToTimespec(deadline)
		if err != nil {
			return err
		}

		timeoutPtr = uintptr(unsafe.Pointer(&unixTimeout))
	}

	for {
		_, _, errno := unix.Syscall6(
			unix.SYS_MQ_TIMEDSEND,
			uintptr(mq.fd),
			uintptr(unsafe.Pointer(&msg[0])),
			uintptr(len(msg)),
			uintptr(priority),
			timeoutPtr,
			0, // Last value unused
		)

		switch errno {
		case 0:
			return nil
		case unix.EINTR:
			continue
		case unix.EBADF:
			return ErrMessageQueueInvalid
		case unix.EMSGSIZE:
			return ErrMessageTooLarge
		case unix.ETIMEDOUT, unix.EAGAIN:
			return ErrMessageQueueFull
		default:
			return errno
		}
	}
}

func (mq *MessageQueue) Send(msg []byte, priority uint) error {
	return mq.commonSend(msg, priority, nil)
}

func (mq *MessageQueue) TimedSend(msg []byte, priority uint, timeout time.Duration) error {
	return mq.commonSend(msg, priority, &timeout)
}

func (mq *MessageQueue) commonReceive(timeout *time.Duration) ([]byte, uint, error) {
	var recvPriority uint
	recvBuf := make([]byte, mq.attributes.MaxMessageSize)

	var timeoutPtr uintptr = 0
	if timeout != nil {
		deadline := time.Now().Add(*timeout)
		unixTimeout, err := unix.TimeToTimespec(deadline)
		if err != nil {
			return nil, 0, err
		}

		timeoutPtr = uintptr(unsafe.Pointer(&unixTimeout))
	}

	for {
		size, _, errno := unix.Syscall6(
			unix.SYS_MQ_TIMEDRECEIVE,
			uintptr(mq.fd),
			uintptr(unsafe.Pointer(&recvBuf[0])),
			uintptr(len(recvBuf)),
			uintptr(recvPriority),
			timeoutPtr,
			0, // Last value unused
		)

		// EINVAL and EMSGSIZE should never occur since we manage those values
		switch errno {
		case 0:
			return recvBuf[0:int(size)], recvPriority, nil
		case unix.EINTR:
			continue
		case unix.EBADF:
			return nil, 0, ErrMessageQueueInvalid
		case unix.ETIMEDOUT, unix.EAGAIN:
			return nil, 0, ErrMessageQueueEmpty
		default:
			return nil, 0, errno
		}
	}
}

func (mq *MessageQueue) Receive() ([]byte, uint, error) {
	return mq.commonReceive(nil)
}

func (mq *MessageQueue) TimedReceive(timeout time.Duration) ([]byte, uint, error) {
	return mq.commonReceive(&timeout)
}
