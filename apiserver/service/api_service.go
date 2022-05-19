package service

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/repository/metadata"
	"goodfs/apiserver/repository/metadata/version"
	"goodfs/apiserver/service/dataserv"
	"goodfs/apiserver/service/objectstream"
	"goodfs/lib/util"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

// LocateFile 根据Hash定位所有分片位置
func LocateFile(hash string) ([]string, bool) {
	//初始化一个消息发送方
	prov, e := global.AmqpConnection.NewProvider()
	if e != nil {
		return nil, false
	}
	defer prov.Close()
	prov.Exchange = "dataServers"
	//初始化一个消息接收方（无交换机直接入队）
	conm, e := global.AmqpConnection.NewConsumer()
	if e != nil {
		return nil, false
	}
	defer conm.Close()
	conm.DeleteUnused = true

	if ch, ok := conm.Consume(); ok {
		//发送定位请求
		for i := 0; i < global.Config.Rs.AllShards(); i++ {
			prov.Publish(amqp.Publishing{
				ReplyTo: conm.QueName,
				Body:    []byte(fmt.Sprintf("%s.%d", hash, i)),
			})
		}
		locates := make([]string, global.Config.Rs.AllShards())
		cnt := 0
		for cnt < global.Config.Rs.AllShards() {
			select {
			case <-time.After(1 * time.Second):
				log.Warnln("Locate message timeout")
				return locates, cnt == global.Config.Rs.AllShards()
			case resp := <-ch:
				cnt++
				shardName := resp.Type
				idx, _ := strconv.Atoi(strings.Split(shardName, ".")[1])
				ip := string(resp.Body)
				locates[idx] = ip
			}
		}
		return locates, cnt == global.Config.Rs.AllShards()
	}
	return nil, false
}

func GetMetaVersion(hash string) (*meta.Version, int32, bool) {
	res, num := version.Find(hash)
	if res == nil {
		return nil, -1, false
	}
	return res, num, true
}

func GetMetaData(name string, ver int32) (*meta.Data, bool) {
	res := metadata.FindByNameAndVerMode(name, metadata.VerMode(ver))
	if res == nil {
		return nil, false
	}
	return res, true
}

func SaveMetadata(md *meta.Data) (int32, error) {
	ver := md.Versions[0]
	metaD := metadata.FindByNameAndVerMode(md.Name, metadata.VerModeNot)
	var verNum int32
	if metaD != nil {
		verNum = version.Add(nil, metaD.Id, ver)
	} else {
		verNum = 0
		var e error
		if metaD, e = metadata.Insert(md); e != nil {
			verNum = version.ErrVersion
		}
	}

	if verNum == version.ErrVersion {
		return -1, ErrInternalServer
	} else {
		return verNum, nil
	}
}

func StoreObject(req *model.PutReq, md *meta.Data) (int32, error) {
	ver := md.Versions[0]

	//文件数据保存
	if req.Locate == nil {
		var e error
		if ver.Locate, e = streamToDataServer(req, ver.Size); e != nil {
			return -1, e
		}
	} else {
		ver.Locate = req.Locate
	}

	//元数据保存
	return SaveMetadata(md)
}

func streamToDataServer(req *model.PutReq, size int64) ([]string, error) {
	//stream to store
	stream, e := dataServerStream(req.FileName, size)
	if e != nil {
		return nil, e
	}

	//digest validation
	if global.Config.EnableHashCheck {
		reader := io.TeeReader(bufio.NewReaderSize(req.Body, 2048), stream)
		hash := util.SHA256Hash(reader)
		if hash != req.Hash {
			log.Infof("Digest of %v validation failure\n", req.Name)
			if e = stream.Commit(false); e != nil {
				log.Errorln(e)
			}
			return nil, ErrInvalidFile
		}
	} else {
		if _, e = io.CopyBuffer(stream, req.Body, make([]byte, 2048)); e != nil {
			if e = stream.Commit(false); e != nil {
				log.Errorln(e)
			}
			return nil, ErrInternalServer
		}
	}

	if e = stream.Commit(true); e != nil {
		log.Errorln(e)
		return nil, ErrServiceUnavailable
	}
	return stream.Locates, e
}

func GetObject(ver *meta.Version) (io.ReadSeekCloser, error) {
	r, e := objectstream.NewRSGetStream(ver.Size, ver.Hash, ver.Locate)
	if e == objectstream.ErrNeedUpdateMeta {
		version.Update(nil, ver)
		e = nil
	}
	return r, e
}

func SelectDataServer(size int) []string {
	ds := dataserv.GetDataServers()
	if len(ds) == 0 {
		return []string{}
	}
	serv := make([]string, size)
	for i := 0; i < global.Config.Rs.AllShards(); i++ {
		if len(ds) >= global.Config.Rs.AllShards()-i {
			ds, serv[i] = global.Balancer.Pop(ds)
		} else {
			serv[i] = global.Balancer.Select(ds)
		}
	}
	return serv
}

func dataServerStream(name string, size int64) (*objectstream.RSPutStream, error) {
	ds := SelectDataServer(global.Config.Rs.AllShards())
	if len(ds) == 0 {
		return nil, ErrServiceUnavailable
	}
	return objectstream.NewRSPutStream(ds, name, size)
}
