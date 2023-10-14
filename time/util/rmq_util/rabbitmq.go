package rmq_util

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	KEY      = "user.event.publish"
	QUEUE    = "user-published-queue"
	CONSUMER = "user-published-consumer"
)

// RabbitMQ stores rabbitmq's connection information
// it also handles disconnection (purpose of URL and QueueName storage)
type RabbitMQ struct {
	URL        string
	Exchange   string
	Conn       *amqp.Connection
	Chann      *amqp.Channel
	Queue      amqp.Queue
	closeChann chan *amqp.Error
	quitChann  chan bool
}

func InitRabbitMQ(config AMQP, shouldAck func([]byte) bool) (*RabbitMQ, error) {
	rmq := &RabbitMQ{
		URL:      config.URL,
		Exchange: config.Exchange,
	}

	err := rmq.load(shouldAck)
	if err != nil {
		return nil, err
	}

	rmq.quitChann = make(chan bool)

	go rmq.handleDisconnect(shouldAck)

	return rmq, err
}

func (rmq *RabbitMQ) load(shouldAck func([]byte) bool) error {
	var err error

	rmq.Conn, err = amqp.Dial(rmq.URL)
	if err != nil {
		return err
	}

	rmq.Chann, err = rmq.Conn.Channel()
	if err != nil {
		return err
	}

	log.Println("connection to rabbitMQ established")

	rmq.closeChann = make(chan *amqp.Error)
	rmq.Conn.NotifyClose(rmq.closeChann)

	// declare exchange if not exist
	err = rmq.Chann.ExchangeDeclare(rmq.Exchange, "direct", true, false, false, false, nil)
	if err != nil {
		return errors.Wrapf(err, "declaring exchange %q", rmq.Exchange)
	}

	args := make(amqp.Table)
	args["x-delayed-type"] = "direct"
	err = rmq.Chann.ExchangeDeclare("delayed", "x-delayed-message", true, false, false, false, args)
	if err != nil {
		return errors.Wrapf(err, "declaring exchange %q", "delayed")
	}

	err = declareConsumer(rmq, shouldAck)
	if err != nil {
		return err
	}

	return nil
}

// declareConsumer declares all queues and bindings for the consumer
func declareConsumer(rmq *RabbitMQ, shouldAck func([]byte) bool) error {
	var err error

	delayedQueue, err := rmq.Chann.QueueDeclare(QUEUE, true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = rmq.Chann.QueueBind(delayedQueue.Name, "user.event.publish", "delayed", false, nil)
	if err != nil {
		return err
	}

	// Set our quality of service_sched.  Since we're sharing 3 consumers on the same
	// channel, we want at least 2 messages in flight.
	err = rmq.Chann.Qos(1, 0, false)
	if err != nil {
		return err
	}

	published, err := rmq.Chann.Consume(
		QUEUE,
		CONSUMER,
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		return err
	}

	go consume(published, shouldAck)

	return nil
}

func consume(ds <-chan amqp.Delivery, shouldAck func([]byte) bool) {
	for {
		//log.Debug("standing by:")

		select {
		case d, ok := <-ds:
			if !ok {
				return
			}
			//log.Infof("consume: %s", string(d.Body))

			//tick := time.Now().UnixMilli()
			if shouldAck(d.Body) {
				d.Ack(false)
			} else {
				d.Nack(true, false)
			}
			//log.Println("#### shouldAck: time consume =", time.Now().UnixMilli()-tick)
		}
	}
}

// Shutdown closes rabbitmq's connection
func (rmq *RabbitMQ) Shutdown() {
	rmq.quitChann <- true

	log.Println("shutting down rabbitMQ's connection...")

	<-rmq.quitChann
}

// handleDisconnect handle a disconnection trying to reconnect every 5 seconds
func (rmq *RabbitMQ) handleDisconnect(shouldAck func([]byte) bool) {
	for {
		select {
		case errChann := <-rmq.closeChann:
			if errChann != nil {
				log.Printf("rabbitMQ disconnection: %v", errChann)
			}
		case <-rmq.quitChann:
			rmq.Conn.Close()
			log.Printf("...rabbitMQ has been shut down")
			rmq.quitChann <- true
			return
		}

		log.Printf("...trying to reconnect to rabbitMQ...")

		time.Sleep(5 * time.Second)

		if err := rmq.load(shouldAck); err != nil {
			log.Printf("rabbitMQ error: %v", err)
		}
	}
}

func (rmq *RabbitMQ) Publish(routingKey string, body []byte) error {
	return rmq.publish(rmq.Exchange, routingKey, body, 0)
}

func (rmq *RabbitMQ) PublishWithDelay(routingKey string, body []byte, delay int64) error {
	return rmq.publish("delayed", routingKey, body, delay)
}

func (rmq *RabbitMQ) publish(exchange string, routingKey string, body []byte, delay int64) error {
	headers := make(amqp.Table)

	log.Printf("publishing to %q %q", routingKey, body)

	if delay != 0 {
		headers["x-delay"] = delay
	}

	return rmq.Chann.Publish(exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
		Headers:      headers,
	})
}
