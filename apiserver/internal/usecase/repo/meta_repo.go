package repo

import (
	"apiserver/internal/entity"
	"apiserver/lib/mongodb"
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MetadataRepo struct {
	*mongo.Collection
}

func NewMetadataRepo(collection *mongo.Collection) *MetadataRepo {
	return &MetadataRepo{collection}
}

//Find 根据自定义条件查找元数据并根据verMode返回版本
func (m *MetadataRepo) Find(filter bson.M, verMode entity.VerMode) (*entity.MetaData, error) {
	var data entity.MetaData
	opt := options.FindOne()
	if verMode == entity.VerModeNot {
		//不查询版本
		opt.SetProjection(bson.M{"versions": 0})
	} else if verMode == entity.VerModeLast {
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
	err := m.FindOne(context.TODO(), filter, opt).Decode(&data)
	return &data, err
}

//FindById 根据Id查找元数据并返回所有版本
func (m *MetadataRepo) FindById(id string) *entity.MetaData {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("Not found document with id %v", id)
		return nil
	}

	data, err := m.Find(bson.M{"_id": oid}, entity.VerModeALL)

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
func (m *MetadataRepo) FindByName(name string) *entity.MetaData {
	return m.FindByNameAndVerMode(name, entity.VerModeALL)
}

//FindByNameAndVerMode 根据文件名查找元数据 verMode筛选版本数据
func (m *MetadataRepo) FindByNameAndVerMode(name string, verMode entity.VerMode) *entity.MetaData {
	data, err := m.Find(bson.M{"name": name}, verMode)
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
func (m *MetadataRepo) FindByHash(hash string) *entity.MetaData {
	var data entity.MetaData
	err := m.FindOne(
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

func (m *MetadataRepo) Insert(data *entity.MetaData) (*entity.MetaData, error) {
	if data.Versions == nil {
		data.Versions = make([]*entity.Version, 0)
	} else {
		tn := time.Now()
		for _, v := range data.Versions {
			v.Ts = tn
		}
	}

	data.CreateTime = time.Now()
	data.UpdateTime = data.CreateTime

	res, err := m.InsertOne(context.TODO(), data)
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	data.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return data, nil
}

func (m *MetadataRepo) Exist(filter bson.M) bool {
	cnt, e := m.CountDocuments(nil, filter)
	if e != nil || cnt == 0 {
		log.Errorln(e)
		return false
	}
	return true
}

// Update 暂时没什么用
// 不允许在这个方法上直接更新versions数组
func (m *MetadataRepo) Update(data *entity.MetaData) error {
	if data == nil || mongodb.GetObjectID(data.Id).IsZero() {
		return errors.New("metadata is nil or id is empty")
	}
	data.UpdateTime = time.Now()
	_, err := m.UpdateByID(context.TODO(), data.Id, bson.M{
		"$set": bson.M{
			"tags": data.Tags,
		},
	})
	return err
}
