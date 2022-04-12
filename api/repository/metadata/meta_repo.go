package metadata

import (
	"context"
	"errors"
	"goodfs/api/model/meta"
	"goodfs/api/repository"
	"goodfs/util"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	//VerModeALL 查询全部版本
	VerModeALL = -128
	//VerModeLast 只查询最后一个版本
	VerModeLast = -2
	// VerModeNot 不查询任何版本
	VerModeNot = -1
)

func Find(filter bson.M, verMode int) (*meta.MetaData, error) {
	collection := repository.GetMongo().Collection("metadata")
	var data meta.MetaData
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

func FindById(id string) *meta.MetaData {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("Not found document with id %v", id)
		return nil
	}

	data, err := Find(bson.M{"_id": oid}, VerModeALL)

	if err == mongo.ErrNoDocuments {
		log.Printf("Not found document with id %v", id)
		return nil
	} else if err != nil {
		log.Print(err)
		return nil
	}
	return data
}

func FindByName(name string) *meta.MetaData {
	return FindByNameAndVerMode(name, VerModeALL)
}

func FindByNameAndVerMode(name string, verMode int) *meta.MetaData {
	data, err := Find(bson.M{"name": name}, verMode)
	if err == mongo.ErrNoDocuments {
		log.Printf("Not found document with name %v", name)
		return nil
	} else if err != nil {
		log.Print(err)
		return nil
	}
	return data
}

func FindByHash(hash string) *meta.MetaData {
	data, err := Find(bson.M{"hash": hash}, VerModeALL)
	if err == mongo.ErrNoDocuments {
		log.Printf("Not found document with hash %v", hash)
		return nil
	} else if err != nil {
		log.Print(err)
		return nil
	}
	return data
}

func Insert(data *meta.MetaData) (*meta.MetaData, error) {
	if data.Versions == nil {
		data.Versions = make([]*meta.MetaVersion, 0)
	} else {
		tn := time.Now()
		for _, v := range data.Versions {
			v.Ts = tn
		}
	}

	data.CreateTime = time.Now()
	data.UpdateTime = data.CreateTime

	collection := repository.GetMongo().Collection("metadata")
	res, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	data.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return data, nil
}

func Exist(filter bson.M) bool {
	collection := repository.GetMongo().Collection("metadata")
	cnt, e := collection.CountDocuments(nil, filter)
	if e != nil || cnt == 0 {
		log.Println(e)
		return false
	}
	return true
}

// Update 暂时没什么用
// 不允许在这个方法上直接更新versions数组
func Update(data *meta.MetaData) error {
	if data == nil || util.GetObjectID(data.Id).IsZero() {
		return errors.New("metadata is nil or id is empty")
	}
	data.UpdateTime = time.Now()
	collection := repository.GetMongo().Collection("metadata")
	_, err := collection.UpdateByID(context.TODO(), data.Id, bson.M{
		"tags": data.Tags,
	})
	return err
}
