package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

const totalRetries = 10
const retryDelay = 15

type session struct {
	*amqp.Connection
	*amqp.Channel
}

// Subscriber represents a single consumer to an exchange
type Subscriber struct {
	ctx          context.Context
	done         context.CancelFunc
	uri          string
	exchange     string
	exchangeType string
	queue        string
	key          string
	ctag         string
	retries      int
	Messages     chan *amqp.Delivery
}

// NewSubscriber creates and initializes a new subscriber
func NewSubscriber(ctx context.Context, done context.CancelFunc, uri, exchange, exchangeType, queue, key, ctag string) (*Subscriber, error) {
	messages := make(chan *amqp.Delivery)
	retries := 0

	s := &Subscriber{
		ctx,
		done,
		uri,
		exchange,
		exchangeType,
		queue,
		key,
		ctag,
		retries,
		messages,
	}

	return s, nil
}

// Consume begins the subscribe and consume loop of a subscruber. Messages will be pumped through the Messages channel.
func (subscriber *Subscriber) Consume() {
	go func() {
		subscriber.subscribeAndConsume()
		subscriber.done()
	}()
}

// Close cleans up a subscriber and it's resources
func (subscriber *Subscriber) Close() {
	close(subscriber.Messages)
}

func (subscriber *Subscriber) redial() chan chan session {
	sessions := make(chan chan session)

	go func() {
		s := make(chan session)
		defer close(sessions)

		for {
			if subscriber.retries > totalRetries {
				log.Fatal("Too many retries")
			}
			if subscriber.retries >= 1 {
				select {
				case <-subscriber.ctx.Done():
					log.Info("Shutting down session factory")
					return
				default:
					break
				}

				log.Info("Waiting to retry")
				time.Sleep(time.Second * retryDelay)
			} else {
				select {
				case sessions <- s:
				case <-subscriber.ctx.Done():
					log.Info("Shutting down session factory")
					return
				}
			}

			subscriber.retries++

			log.WithFields(log.Fields{
				"uri": subscriber.uri,
			}).Info("Dialing")
			conn, err := amqp.Dial(subscriber.uri)
			if err != nil {
				log.Error(err)
				continue
			}

			log.Info("Opening channel")
			ch, err := conn.Channel()
			if err != nil {
				log.Error(err)
				continue
			}

			log.WithFields(log.Fields{
				"exchange": subscriber.exchange,
				"type":     subscriber.exchangeType,
			}).Info("Declaring exchange")
			err = ch.ExchangeDeclare(
				subscriber.exchange,
				subscriber.exchangeType,
				true,  // durable
				false, // delete when complete
				false, // internal
				false, // noWait
				nil,   // arguments
			)
			if err != nil {
				log.Error(err)
				continue
			}

			// If we get here, we are connected succesfully, reset retries
			subscriber.retries = 0

			select {
			case s <- session{conn, ch}:
			case <-subscriber.ctx.Done():
				log.Info("Shutting down new session")
			}
		}
	}()

	return sessions
}

func (subscriber *Subscriber) subscribeAndConsume() {
	for session := range subscriber.redial() {
		sub := <-session

		log.WithFields(log.Fields{
			"queue": subscriber.queue,
		}).Infoln("Declaring queue")
		if _, err := sub.QueueDeclare(
			subscriber.queue,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // noWait
			nil,   // arguments)
		); err != nil {
			log.Error(err)
			return
		}

		log.WithFields(log.Fields{
			"queue":    subscriber.queue,
			"key":      subscriber.key,
			"exchange": subscriber.exchange,
		}).Infoln("Binding queue")
		if err := sub.QueueBind(
			subscriber.queue,
			subscriber.key,
			subscriber.exchange,
			false,
			nil,
		); err != nil {
			log.Error(err)
			return
		}

		log.WithFields(log.Fields{
			"queue": subscriber.queue,
			"ctag":  subscriber.ctag,
		}).Infoln("Consuming")
		deliveries, err := sub.Consume(
			subscriber.queue,
			subscriber.ctag,
			false, // noAck
			false, // exclusive
			false, // noLocal
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			log.Error(err)
			return
		}

		for msg := range deliveries {
			subscriber.Messages <- &msg
			sub.Ack(msg.DeliveryTag, false)
		}
	}
}
