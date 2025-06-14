package publisher

import (
	"time"

	"github.com/streadway/amqp"
)

// session composes an amqp.Connection with an amqp.Channel
type session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close tears the connection down, taking the channel with it.
func (s session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

type Message struct {
	Exchange    string
	Type        string
	Body        []byte
	Sub         session
	DeliveryTag uint64
}

func (m *Message) Success() {
	m.Sub.Ack(m.DeliveryTag, false)
}

func (m *Message) Fail() {
	// default redelivery time
	time.Sleep(time.Second * 3)
	m.Sub.Nack(m.DeliveryTag, false, true)
}
