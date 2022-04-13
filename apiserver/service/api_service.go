package service

import (
	"encoding/json"
	"goodfs/apiserver/config"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/repository/metadata"
	"goodfs/apiserver/repository/metadata/version"
	"goodfs/apiserver/service/objectstream"
	"goodfs/apiserver/service/selector"
	"goodfs/util"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

var (
	balancer = selector.NewSelector(config.SelectStrategy)
)

func LocateFile(name string) (string, bool) {
	//初始化一个消息发送方
	prov, e := global.AmqpConnection.NewProvider()
	if e != nil {
		return "", false
	}
	defer prov.Close()
	prov.Exchange = "dataServers"
	//初始化一个消息接收方（无交换机直接入队）
	conm, e := global.AmqpConnection.NewConsumer()
	if e != nil {
		return "", false
	}
	defer conm.Close()
	conm.DeleteUnused = true

	if ch, ok := conm.Consume(); ok {
		//发送定位请求
		jn, e := json.Marshal(name)
		if e != nil {
			return "", false
		}
		prov.Publish(amqp.Publishing{
			ReplyTo: conm.QueName,
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
	return "", false
}

func GetMetaData(name string, ver int) (*meta.MetaData, bool) {
	res := metadata.FindByNameAndVerMode(name, ver)
	if res == nil {
		return nil, false
	}
	return res, true
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
