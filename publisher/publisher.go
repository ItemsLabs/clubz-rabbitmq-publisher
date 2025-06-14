package publisher

import (
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/lib/pq"
	"github.com/streadway/amqp"
)

const (
	NotifyTimeout = 3 * time.Second
)

const (
	getEventsSql = `
		select id,
		       exchange,
		       type,
		       data
		  from amqp_events
		 order by id
		 limit 100`

	deleteEventsSql = `
		delete
		  from amqp_events
		 where id = ANY($1)`
)

type EventPublisher struct {
	Connector
	db       *sql.DB
	listener *pq.Listener

	url       string
	exchanges []string
	queue     string
	messages  chan *Message
}

func (ev *EventPublisher) Start() {
	ev.messages = make(chan *Message)
	go func() {
		ev.startPublishing(ev.connectToExchange(ev.url, ev.exchanges))
		defer close(ev.messages)
	}()
}

func (ev *EventPublisher) Stop() {
	ev.Close()
}

// subscribe consumes deliveries from an exclusive queue from a fanout exchange and sends to the application specific messages chan.
func (ev *EventPublisher) startPublishing(sessions chan chan session) {
	for session := range sessions {
		pub, alive := <-session

		if !alive {
			continue
		}

		for {
			err := ev.processMessage(pub)
			if err != nil {
				break
			}
		}

		close(session)
	}
}

func (ev *EventPublisher) runInsideTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := ev.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("cannot start db transaction")
		return err
	}

	// Rollback the transaction on panics in the action. Don't swallow the
	// panic, though, let it propagate.
	defer func(tx *sql.Tx) {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}(tx)

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (ev *EventPublisher) processMessage(sess session) error {
	var (
		evID       int
		evExchange string
		evType     string
		evData     string
	)

	rows, err := ev.db.Query(getEventsSql)
	if err != nil {
		logrus.WithError(err).Error("error during query row")

		return nil
	}
	defer func() {
		_ = rows.Close()
	}()

	var processedEvents []int
	for rows.Next() {
		if err = rows.Scan(&evID, &evExchange, &evType, &evData); err != nil {
			logrus.WithError(err).Error("error during scan row")
			return err
		}

		// publish event to amqp
		err = ev.publishEvent(
			sess, &AMQPEvent{
				ID:       evID,
				Exchange: evExchange,
				Type:     evType,
				Data:     evData,
			},
		)

		if err != nil {
			logrus.WithError(err).Error("error during publish event to rabbitmq")
			return err
		}

		processedEvents = append(processedEvents, evID)
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return err
	}

	// wait a bit for next events
	if len(processedEvents) == 0 {
		// by default we should catch notify event and start next processing step
		// if event is not emitted during "NotifyTimeout" then go to next processing step immediately
		select {
		case <-time.After(NotifyTimeout):
			break
		case <-ev.listener.Notify:
			break
		}
	} else {
		_, err = ev.db.Exec(deleteEventsSql, pq.Array(processedEvents))
		if err != nil {
			logrus.WithError(err).Error("cannot delete processed amqp_events")
			return err
		}
	}

	return nil
}

func (ev *EventPublisher) publishEvent(sess session, event *AMQPEvent) error {
	// try to publish into channel
	logrus.WithField("id", event.ID).Info("publish event")

	err := sess.Publish(
		event.Exchange, "", false, false, amqp.Publishing{
			// use persistent delivery mode (mode = 2) for maximum consistency
			DeliveryMode: 2,
			Type:         event.Type,
			Body:         []byte(event.Data),
		},
	)
	if err != nil {
		logrus.WithError(err).Error("cannot publish")
		return err
	}

	return nil
}

func NewEventPublisher(url string, exchanges []string, db *sql.DB, listener *pq.Listener) *EventPublisher {
	publisher := &EventPublisher{
		db:        db,
		listener:  listener,
		url:       url,
		exchanges: exchanges,
	}
	publisher.Init()
	return publisher
}
