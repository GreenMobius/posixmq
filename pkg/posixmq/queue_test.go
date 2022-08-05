package posixmq_test

import (
	"testing"

	"github.com/GreenMobius/posixmq/pkg/posixmq"
	"golang.org/x/sys/unix"
)

func TestMessageQueueOpenClose(t *testing.T) {
	mq, err := posixmq.Open("posixmq_test", unix.O_CREAT, posixmq.MessageQueueAttributes{
		MaxQueueSize:   posixmq.MessageQueueMaxQueueSize,
		MaxMessageSize: posixmq.MessageQueueMaxMessageSize,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mq.Close(); err != nil {
		t.Fatal(err)
	}

	if err := mq.Unlink(); err != nil {
		t.Fatal(err)
	}
}
