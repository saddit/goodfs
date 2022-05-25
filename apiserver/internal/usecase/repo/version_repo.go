package repo

import (
	"apiserver/internal/entity"
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ErrVersion int32 = -1
)

type VersionRepo struct {
	*mongo.Collection
}

func NewVersionRepo(collection *mongo.Collection) *VersionRepo {
	return &VersionRepo{collection}
}

//Find 根据hash查找版本，返回版本以及版本号
func (v *VersionRepo) Find(hash string) (*entity.Version, int32) {
	res := struct {
		Index    int32             `bson:"index"`
		Versions []*entity.Version `bson:"versions"`
	}{}
	if e := v.FindOne(
		nil,
		bson.M{"versions.hash": hash},
		options.FindOne().SetProjection(bson.M{
			"index":      bson.M{"$indexOfArray": bson.A{"$versions.hash", hash}},
			"versions.$": 1,
		}),
	).Decode(&res); e != nil {
		log.Errorln(e)
		return nil, ErrVersion
	}

	if len(res.Versions) > 0 {
		return res.Versions[0], res.Index
	} else {
		return nil, res.Index
	}
}

//Update updating locate and setting ts to now
func (v *VersionRepo) Update(ctx context.Context, ver *entity.Version) bool {
	res, e := v.UpdateOne(ctx, bson.M{
		"versions.hash": ver.Hash,
	}, bson.M{
		"$set": bson.M{
			"versions.$.locate": ver.Locate,
			"versions.$.ts":     time.Now(),
		},
	})
	if e != nil {
		log.Errorf("Error when update version %v: %v", ver.Hash, e)
	}
	return res.ModifiedCount == 1
}

//Add 为metadata添加一个版本，添加到版本数组的末尾，版本号为数组序号
//返回对应版本号,如果失败返回ErrVersion -1
func (v *VersionRepo) Add(ctx context.Context, id string, ver *entity.Version) int32 {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Warnf("id error %v", id)
		return ErrVersion
	}

	ver.Ts = time.Now()
	data := struct {
		LenOfVersion int32 `bson:"lenOfVersion"`
	}{}

	//returns the pre-modified version of the document
	err = v.FindOneAndUpdate(ctx, bson.M{
		"_id": oid,
	}, bson.M{
		"$set": bson.M{
			"update_time": time.Now(),
		},
		"$push": bson.M{
			"versions": ver,
		},
	}, options.FindOneAndUpdate().SetProjection(bson.M{
		"lenOfVersion": bson.M{"$size": "$versions"},
		"_id":          0,
	})).Decode(&data)

	if err != nil {
		log.Errorln(err)
		return ErrVersion
	}

	return data.LenOfVersion
}

func (v *VersionRepo) Delete(ctx context.Context, id string, ver *entity.Version) error {
	if ctx == nil {
		ctx = context.Background()
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Warnf("id error %v", id)
		return nil
	}

	res, err := v.UpdateOne(ctx, bson.M{
		"_id": oid,
	}, bson.M{
		"$set": bson.M{
			"versions.$[elem].hash": "",
			"versions.$[elem].size": 0,
		},
	}, options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{
			"elem.hash": ver.Hash,
		}},
	}).SetHint("metadata_versions_hash"))

	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return errors.New("Delete fail")
	}

	return nil
}
