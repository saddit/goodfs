package rabbitmq

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type AMQP interface {
	Send(queue string, body interface{})
	Publish(exchange string, body interface{}) error
	Consume(name string) <-chan amqp.Delivery
	NewChannel() error
}

type RabbitMQ struct {
	channel  *amqp.Channel    //channel
	conn     *amqp.Connection //connection
	Name     string           //queue name
	exchange string
}

func New(s string) *RabbitMQ {
	conn, e := amqp.Dial(s)
	if e != nil {
		log.Panicf("Dail %v failed, %v", s, e)
	}

	ch, e := conn.Channel()
	if e != nil {
		panic(e)
	}

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.conn = conn
	return mq
}

func NewConn(s string) *amqp.Connection {
	conn, e := amqp.Dial(s)
	if e != nil {
		log.Panicf("Dail %v failed, %v", s, e)
	}
	return conn
}

/*
	创建一个直接消费队列
	Not durable, auto delete
*/
func (q *RabbitMQ) DeclareQueue(name string) {
	que, e := q.channel.QueueDeclare(
		name,  // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if e != nil {
		panic(e)
	}
	q.Name = que.Name
}

/*
	绑定交换机与当前声明的队列
*/
func (q *RabbitMQ) CreateBind(exchange, qName string) {
	e := q.channel.QueueBind(
		qName,    // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil)
	if e != nil {
		panic(e)
	}
	q.exchange = exchange
}

func (q *RabbitMQ) Bind(exchange string) {
	q.CreateBind(exchange, q.Name)
}

/*
	直接向队列发送消息
*/
func (q *RabbitMQ) Send(queue string, body interface{}) {
	str, e := json.Marshal(body)
	if e != nil {
		panic(e)
	}
	e = q.channel.Publish("",
		queue,
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    str,
		})
	if e != nil {
		panic(e)
	}
}

func (q *RabbitMQ) Publish(exchange string, body interface{}) error {
	str, e := json.Marshal(body)
	if e != nil {
		return e
	}
	e = q.channel.Publish(exchange,
		"", //route key
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    str,
		})
	if e != nil {
		return e
	}
	return nil
}

func (q *RabbitMQ) Consume(name string) <-chan amqp.Delivery {
	c, e := q.channel.Consume(
		q.Name,
		name, //consumer name
		true, //auto ack
		false,
		false,
		false,
		nil,
	)
	if e != nil {
		panic(e)
	}
	return c
}

func (q *RabbitMQ) NewChannel() error {
	chl, e := q.conn.Channel()
	if e != nil {
		return e
	}
	q.channel = chl
	return nil
}

func (q *RabbitMQ) Close() {
	err := q.channel.Close()
	err = q.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (q *RabbitMQ) CloseChannel() {
	err := q.channel.Close()
	if err != nil {
		panic(err)
	}
}

func (q *RabbitMQ) IsClose() bool {
	return q.conn.IsClosed()
}
