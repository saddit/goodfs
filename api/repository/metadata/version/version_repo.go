package version

import (
	"context"
	"errors"
	"goodfs/api/model/meta"
	"goodfs/api/repository"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ErrVersion = -1
)

// Add 为metadata添加一个版本，添加到版本数组的末尾，版本号为数组序号
//返回对应版本号,如果失败返回ErrVersion -1
func Add(ctx context.Context, id string, ver *meta.MetaVersion) int {
	if ctx == nil {
		ctx = context.Background()
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("id error %v", id)
		return ErrVersion
	}

	ver.Ts = time.Now()
	collection := repository.GetMongo().Collection("metadata")
	data := struct {
		LenOfVersion int `bson:"lenOfVersion"`
	}{}

	//returns the pre-modified version of the document
	err = collection.FindOneAndUpdate(ctx, bson.M{
		"_id": oid,
	}, bson.M{
		"$set": bson.M {
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
		log.Println(err)
		return ErrVersion
	}

	return data.LenOfVersion
}

func Delete(ctx context.Context, id string, ver *meta.MetaVersion) error {
	if ctx == nil {
		ctx = context.Background()
	}

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("id error %v", id)
		return nil
	}

	collection := repository.GetMongo().Collection("metadata")

	res, err := collection.UpdateOne(ctx, bson.M{
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
