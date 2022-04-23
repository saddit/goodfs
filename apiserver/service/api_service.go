package service

import (
	"bufio"
	"encoding/json"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/repository/metadata"
	"goodfs/apiserver/repository/metadata/version"
	"goodfs/apiserver/service/objectstream"
	"goodfs/lib/util"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
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
			Body:    jn,
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

func GetMetaVersion(hash string) (*meta.MetaVersion, int32, bool) {
	res, num := version.Find(hash)
	if res == nil {
		return nil, -1, false
	}
	return res, num, true
}

func SendExistingSyncMsg(hv []byte, typing model.SyncTyping) error {
	//初始化一个消息发送方
	prov, e := global.AmqpConnection.NewProvider()
	if e != nil {
		return e
	}
	defer prov.Close()
	prov.Exchange = "existSync"
	if prov.Publish(amqp.Publishing{Body: hv, Type: string(typing)}) {
		return nil
	}
	return ErrServiceUnavailable
}

func GetMetaData(name string, ver int32) (*meta.MetaData, bool) {
	res := metadata.FindByNameAndVerMode(name, metadata.VerMode(ver))
	if res == nil {
		return nil, false
	}
	return res, true
}

func StoreObject(req *model.PutReq, md *meta.MetaData) (int32, error) {
	ver := md.Versions[0]

	//文件数据保存
	if req.Locate == "" {
		var e error
		if ver.Locate, e = streamToDataServer(req, ver.Size); e != nil {
			return -1, e
		}
	} else {
		ver.Locate = req.Locate
	}

	//元数据保存
	metaD := metadata.FindByNameAndVerMode(md.Name, metadata.VerModeNot)
	var verNum int32
	if metaD != nil {
		verNum = version.Add(nil, metaD.Id, ver)
	} else {
		verNum = 0
		metaD, _ = metadata.Insert(md)
	}

	if verNum == version.ErrVersion {
		return -1, ErrInternalServer
	} else {
		return verNum, nil
	}
}

func streamToDataServer(req *model.PutReq, size int64) (string, error) {
	//stream to store
	stream, e := dataServerStream(req.FileName, size)
	if e != nil {
		return "", e
	}

	//digest validation
	reader := io.TeeReader(bufio.NewReaderSize(req.Body, 2048), stream)
	hash := util.SHA256Hash(reader)
	if hash != req.Hash {
		if e = stream.Commit(false); e != nil {
			log.Println(e)
		}
		return "", ErrInvalidFile
	}

	if e = stream.Commit(true); e != nil {
		log.Println(e)
		return "", ErrServiceUnavailable
	}
	return stream.Locate, e
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

func dataServerStream(name string, size int64) (*objectstream.PutStream, error) {
	ds := GetDataServers()
	if len(ds) == 0 {
		return nil, ErrServiceUnavailable
	}
	serv := global.Balancer.Select(ds)
	return objectstream.NewPutStream(serv, name, size)
}
