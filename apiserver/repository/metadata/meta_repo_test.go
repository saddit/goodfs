package metadata

import (
	"encoding/json"
	"goodfs/apiserver/model/meta"
	"goodfs/apiserver/repository/metadata/version"
	"goodfs/util"
	"math/rand"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randStringRunes(n int) string {
	rand.Seed(time.Now().Unix())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestFindById(t *testing.T) {
	res, e := Find(bson.M{"_id": util.GetObjectID("624bdb6b2266000007007824")}, VerModeLast)
	if e != nil {
		t.Error("Not found", e)
	} else {
		j, e := json.MarshalIndent(res, "  ", "  ")
		if e != nil {
			t.Error(e)
			return
		}
		t.Logf("Found data\n %v", string(j))
	}
}

func TestFindByName(t *testing.T) {
	res := FindByNameAndVerMode("Vivi3.mp4", VerModeLast)
	if res == nil {
		t.Error("Not found")
	} else {
		j, e := json.MarshalIndent(res, "  ", "  ")
		if e != nil {
			t.Error(e)
			return
		}
		t.Logf("Found data\n %v", string(j))
	}
}

func TestInsert(t *testing.T) {
	res, err := Insert(&meta.MetaData{
		Name: randStringRunes(10) + ".txt",
		Tags: []string{"text"},
		Versions: []*meta.MetaVersion{{
			Hash:   randStringRunes(32),
			Locate: "0.0.0.0",
			Size:   rand.Int63n(9999999),
		}},
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("Success with id %v", res.Id)
	}
}

func TestAddVersion(t *testing.T) {
	verCode := version.Add(nil, "624c0c0cf0a7aab7f5628498", &meta.MetaVersion{
		Hash:   randStringRunes(32),
		Size:   rand.Int63n(999999),
		Locate: "0.0.0.0",
	})
	if verCode == version.ErrVersion {
		t.Error("Add version fail")
		return
	}
	t.Logf("After add new version, verCode is %v", verCode)
}

func TestDelVer(t *testing.T) {
	res := FindByName("Vivi3.mp4")
	if res == nil {
		t.Error("Not found")
		return
	}
	j, e := json.MarshalIndent(res, "  ", "  ")
	if e != nil {
		t.Error(e)
		return
	}
	t.Logf("Found data\n %v", string(j))
	e = version.Delete(nil, res.Id, res.Versions[0])
	if e != nil {
		t.Error(e)
	}
}
