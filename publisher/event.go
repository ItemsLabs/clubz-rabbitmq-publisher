package publisher

type AMQPEvent struct {
	ID       int
	Exchange string
	Type     string
	Data     string
}
