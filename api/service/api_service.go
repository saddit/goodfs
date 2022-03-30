package service

import (
	"encoding/json"
	"errors"
	"goodfs/api/config"
	"goodfs/api/objectstream"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

var (
	ErrorServiceUnavailable = errors.New("DataServer unavailable")
	ErrorInternalServer     = errors.New("Internal server error")
	amqpConn, _             = amqp.Dial(config.AmqpAddress)
)

func LocateFile(name string) (string, bool) {
	chl, e := amqpConn.Channel()
	if e != nil {
		return "", false
	}
	defer chl.Close()
	//reply 队列
	que, e := chl.QueueDeclare("", false, true, false, false, nil)
	if e != nil {
		return "", false
	}
	//监听 reply
	ch, e := chl.Consume(que.Name, "", true, false, false, false, nil)
	if e != nil {
		return "", false
	}
	//发送定位请求
	jn, e := json.Marshal(name)
	if e != nil {
		return "", false
	}
	chl.Publish("dataServers", "", false, false, amqp.Publishing{
		ReplyTo: que.Name,
		Body:    []byte(jn),
	})

	select {
	case <-time.After(1 * time.Second):
		log.Println("Locate message timeout")
		return "", false
	case resp := <-ch:
		s, ok := strconv.Unquote(string(resp.Body))
		return s, ok == nil
	}
}

func StoreObject(reader io.Reader, name string) error {
	stream, e := dataServerStream(name)
	if e != nil {
		return e
	}
	io.CopyBuffer(stream, reader, make([]byte, 2048))
	return stream.Close()
}

func GetObject(ip string, name string) (io.Reader, error) {
	return objectstream.NewGetStream(ip, name)
}

func dataServerStream(name string) (io.WriteCloser, error) {
	serv, ok := RandomDataServer()
	if !ok {
		return nil, ErrorServiceUnavailable
	}
	return objectstream.NewPutStream(serv, name), nil
}
