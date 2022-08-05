package posixmq_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/GreenMobius/posixmq/pkg/posixmq"
	"golang.org/x/sys/unix"
)

func TestMessageQueueSendReceive(t *testing.T) {
	mq, err := posixmq.Open("posixmq_test", unix.O_CREAT|unix.O_RDWR, posixmq.MessageQueueAttributes{
		MaxQueueSize:   posixmq.MessageQueueMaxQueueSize,
		MaxMessageSize: posixmq.MessageQueueMaxMessageSize,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mq.Unlink()

	message := []byte("Hello world!")

	if err := mq.Send(message, 0); err != nil {
		t.Fatal(err)
	}

	response, _, err := mq.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(message, response) {
		t.Fatalf("Expected %v\nReceived %v", message, response)
	}

	if err := mq.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMessageQueueSendReceiveTooLarge(t *testing.T) {
	mq, err := posixmq.Open("posixmq_test", unix.O_CREAT|unix.O_RDWR, posixmq.MessageQueueAttributes{
		MaxQueueSize:   posixmq.MessageQueueMaxQueueSize,
		MaxMessageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mq.Unlink()

	message := []byte("Long message!")

	if err := mq.Send(message, 0); !errors.Is(err, posixmq.ErrMessageTooLarge) {
		t.Fatal(err)
	}

	if err := mq.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMessageQueueSendReceiveFullEmpty(t *testing.T) {
	mq, err := posixmq.Open("posixmq_test", unix.O_CREAT|unix.O_RDWR|unix.O_NONBLOCK, posixmq.MessageQueueAttributes{
		MaxQueueSize:   1,
		MaxMessageSize: posixmq.MessageQueueMaxMessageSize,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mq.Unlink()

	_, _, err = mq.Receive()
	if !errors.Is(err, posixmq.ErrMessageQueueEmpty) {
		t.Fatalf("Expected %v\nReceived %v", posixmq.ErrMessageQueueEmpty, err)
	}

	message := []byte("Testing some limits!")
	if err := mq.Send(message, 0); err != nil {
		t.Fatal(err)
	}

	if err := mq.Send(message, 0); !errors.Is(err, posixmq.ErrMessageQueueFull) {
		t.Fatalf("Expected %v\nReceived %v", posixmq.ErrMessageQueueFull, err)
	}

	response, _, err := mq.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(message, response) {
		t.Fatalf("Expected %v\nReceived %v", message, response)
	}

	if err := mq.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMessageQueueTimedSendReceive(t *testing.T) {
	timeoutDelay := 100 * time.Millisecond

	mq, err := posixmq.Open("posixmq_test", unix.O_CREAT|unix.O_RDWR, posixmq.MessageQueueAttributes{
		MaxQueueSize:   1,
		MaxMessageSize: posixmq.MessageQueueMaxMessageSize,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer mq.Unlink()

	message := []byte("Time-sensitive message!")
	if err := mq.TimedSend(message, 0, timeoutDelay); err != nil {
		t.Fatal(err)
	}

	target := time.Now().Add(timeoutDelay)
	if err := mq.TimedSend(message, 0, timeoutDelay); !errors.Is(err, posixmq.ErrMessageQueueFull) {
		t.Fatalf("Expected %v\nReceived %v", posixmq.ErrMessageQueueFull, err)
	}

	if time.Now().Before(target) {
		t.Fatal("TimedSend timed out early")
	}

	response, _, err := mq.TimedReceive(timeoutDelay)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = mq.TimedReceive(timeoutDelay)
	if !errors.Is(err, posixmq.ErrMessageQueueEmpty) {
		t.Fatalf("Expected %v\nReceived %v", posixmq.ErrMessageQueueEmpty, err)
	}

	if !reflect.DeepEqual(message, response) {
		t.Fatalf("Expected %v\nReceived %v", message, response)
	}

	if err := mq.Close(); err != nil {
		t.Fatal(err)
	}
}
