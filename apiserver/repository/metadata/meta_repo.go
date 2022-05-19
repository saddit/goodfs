package metadata

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"goodfs/apiserver/config"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/repository"
	"goodfs/lib/util"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VerMode int32

const (
	//VerModeALL 查询全部版本
	VerModeALL VerMode = -128
	//VerModeLast 只查询最后一个版本
	VerModeLast VerMode = -2
	// VerModeNot 不查询任何版本
	VerModeNot VerMode = -1
)

//Find 根据自定义条件查找元数据并根据verMode返回版本
func Find(filter bson.M, verMode VerMode) (*meta.Data, error) {
	collection := repository.GetMongo().Collection(config.MetadataMongo)
	var data meta.Data
	opt := options.FindOne()
	if verMode == VerModeNot {
		//不查询版本
		opt.SetProjection(bson.M{"versions": 0})
	} else if verMode == VerModeLast {
		//只查询最新版本
		opt.SetProjection(bson.M{
			"versions": bson.M{"$slice": -1},
		})
	} else if verMode >= 0 {
		//查询指定版本
		opt.SetProjection(bson.M{
			"$slice": bson.A{
				"versions", verMode, 1,
			},
		})
	}
	err := collection.FindOne(context.TODO(), filter, opt).Decode(&data)
	return &data, err
}

//FindById 根据Id查找元数据并返回所有版本
func FindById(id string) *meta.Data {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("Not found document with id %v", id)
		return nil
	}

	data, err := Find(bson.M{"_id": oid}, VerModeALL)

	if err == mongo.ErrNoDocuments {
		log.Infof("Not found document with id %v\n", id)
		return nil
	} else if err != nil {
		log.Errorln(err)
		return nil
	}
	return data
}

//FindByName 根据文件名查找元数据并返回所有版本
func FindByName(name string) *meta.Data {
	return FindByNameAndVerMode(name, VerModeALL)
}

//FindByNameAndVerMode 根据文件名查找元数据 verMode筛选版本数据
func FindByNameAndVerMode(name string, verMode VerMode) *meta.Data {
	data, err := Find(bson.M{"name": name}, verMode)
	if err == mongo.ErrNoDocuments {
		log.Infof("Not found document with name %v", name)
		return nil
	} else if err != nil {
		log.Print(err)
		return nil
	}
	return data
}

//FindByHash 按照版本的Hash值查找元数据 只返回一个版本
func FindByHash(hash string) *meta.Data {
	collection := repository.GetMongo().Collection(config.MetadataMongo)
	var data meta.Data
	err := collection.FindOne(
		nil,
		bson.M{"versions.hash": hash},
		options.FindOne().SetProjection(bson.M{
			"name":        1,
			"tags":        1,
			"create_time": 1,
			"update_time": 1,
			"versions.$":  1,
		}),
	).Decode(&data)

	if err == mongo.ErrNoDocuments {
		log.Infof("Not found document with hash %v", hash)
		return nil
	} else if err != nil {
		log.Print(err)
		return nil
	}

	return &data
}

func Insert(data *meta.Data) (*meta.Data, error) {
	if data.Versions == nil {
		data.Versions = make([]*meta.Version, 0)
	} else {
		tn := time.Now()
		for _, v := range data.Versions {
			v.Ts = tn
		}
	}

	data.CreateTime = time.Now()
	data.UpdateTime = data.CreateTime

	collection := repository.GetMongo().Collection(config.MetadataMongo)
	res, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	data.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return data, nil
}

func Exist(filter bson.M) bool {
	collection := repository.GetMongo().Collection(config.MetadataMongo)
	cnt, e := collection.CountDocuments(nil, filter)
	if e != nil || cnt == 0 {
		log.Errorln(e)
		return false
	}
	return true
}

// Update 暂时没什么用
// 不允许在这个方法上直接更新versions数组
func Update(data *meta.Data) error {
	if data == nil || util.GetObjectID(data.Id).IsZero() {
		return errors.New("metadata is nil or id is empty")
	}
	data.UpdateTime = time.Now()
	collection := repository.GetMongo().Collection(config.MetadataMongo)
	_, err := collection.UpdateByID(context.TODO(), data.Id, bson.M{
		"$set": bson.M{
			"tags": data.Tags,
		},
	})
	return err
}
