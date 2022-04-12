package service

import (
	"encoding/json"
	"goodfs/api/config"
	"goodfs/api/model/meta"
	"goodfs/api/repository/metadata"
	"goodfs/api/repository/metadata/version"
	"goodfs/api/service/objectstream"
	"goodfs/api/service/selector"
	"goodfs/lib/rabbitmq"
	"goodfs/util"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

var (
	amqpConn = rabbitmq.NewConn(config.AmqpAddress)
	balancer = selector.NewSelector(config.SelectStrategy)
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

func GetMetaVersion(name string, ver int) (*meta.MetaVersion, bool) {
	res := metadata.FindByNameAndVerMode(name, ver)
	// log.Print(res.Name, res.Id)
	if res == nil || res.Versions == nil || len(res.Versions) == 0 {
		return nil, false
	}
	return res.Versions[0], true
}

func StoreObject(reader io.Reader, name string, ver *meta.MetaVersion) (int, error) {
	if ext, ok := util.GetFileExt(name, true); ok {
		//stream to store
		stream, e := dataServerStream(ver.Hash + ext)
		if e != nil {
			return -1, e
		}
		io.CopyBuffer(stream, reader, make([]byte, 2048))

		//bolck by chan
		e = stream.Close()
		if e != nil {
			return -1, ErrServiceUnavailable
		}

		//store meta data
		ver.Locate = stream.Locate
		metaD := metadata.FindByNameAndVerMode(name, metadata.VerModeNot)
		var verNum int
		if metaD != nil {
			verNum = version.Add(nil, metaD.Id, ver)
		} else {
			verNum = 0
			metaD, e = metadata.Insert(&meta.MetaData{
				Name:     name,
				Versions: []*meta.MetaVersion{ver},
			})
		}

		//meta data save error
		if verNum == version.ErrVersion {
			go objectstream.DeleteObject(name, ver)
			return -1, ErrInternalServer
		} else {
			return verNum, nil
		}
	}
	return -1, ErrBadRequest
}

func GetObject(name string, ver *meta.MetaVersion) (io.Reader, error) {
	if !IsAvailable(ver.Locate) {
		log.Printf("%v server is unavailable", ver.Locate)
		return nil, ErrServiceUnavailable
	}
	if ext, ok := util.GetFileExt(name, true); ok {
		return objectstream.NewGetStream(ver.Locate, ver.Hash+ext)
	}
	return nil, ErrBadRequest
}

func dataServerStream(name string) (*objectstream.PutStream, error) {
	ds := GetDataServers()
	if len(ds) == 0 {
		return nil, ErrServiceUnavailable
	}
	serv := balancer.Select(ds)
	return objectstream.NewPutStream(serv, name), nil
}
