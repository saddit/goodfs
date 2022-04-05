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

func find(filter interface{}) (*meta.MetaData, error) {
	collection := repository.GetMongo().Collection("metadata")
	var data meta.MetaData
	err := collection.FindOne(context.TODO(), filter).Decode(&data)
	return &data, err
}

func FindById(id string) *meta.MetaData {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("Not found document with id %v", id)
		return nil
	}

	data, err := find(bson.M{"_id": oid})

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
	data, err := find(bson.M{"name": name})
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
	data, err := find(bson.M{"hash": hash})
	if err == mongo.ErrNoDocuments {
		log.Printf("Not found document with hash %v", hash)
		return nil
	} else if err != nil {
		log.Print(err)
		return nil
	}
	return data
}

func InsertNew(data *meta.MetaData) (*meta.MetaData, error) {
	if data.Versions == nil {
		data.Versions = make([]meta.MetaVersion, 0)
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

func UpdateBasic(data *meta.MetaData) error {
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

func AddVersion(id string, ver *meta.MetaVersion) *meta.MetaData {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("id error %v", id)
		return nil
	}

	ver.Ts = time.Now()
	collection := repository.GetMongo().Collection("metadata")
	var data meta.MetaData
	err = collection.FindOneAndUpdate(context.TODO(), bson.M{
		"_id": oid,
	}, bson.M{
		"$push": bson.M{
			"versions": ver,
		},
	}).Decode(&data)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &data
}

func DeleteVersion(id string, ver *meta.MetaVersion) error {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Printf("id error %v", id)
		return nil
	}

	collection := repository.GetMongo().Collection("metadata")
	res, err := collection.UpdateOne(context.TODO(), bson.M{
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
