package metadata

import (
	"encoding/json"
	"goodfs/api/model/meta"
	"math/rand"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestFindById(t *testing.T) {
	res := FindById("624bdb6b2266000007007824")
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

func TestFindByName(t *testing.T) {
	res := FindByName("Vivi3.mp4")
	if res == nil {
		t.Error("Not found")
	} else {
		t.Logf("Found data\n %v", res)
	}
}

func TestInsert(t *testing.T) {
	res, err := InsertNew(&meta.MetaData{
		Name: randStringRunes(10) + ".txt",
		Tags: []string{"text"},
		Versions: []meta.MetaVersion{{
			Hash:   randStringRunes(32),
			Locate: "0.0.0.0",
			Size:   rand.Int31n(9999999),
		}},
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("Success with id %v", res.Id)
	}
}

func TestAddVersion(t *testing.T) {
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
	res = AddVersion(res.Id, &meta.MetaVersion{
		Hash:   randStringRunes(32),
		Size:   rand.Int31n(999999),
		Locate: "0.0.0.0",
	})
	if res == nil {
		t.Error("Add version fail")
		return
	}
	j, e = json.MarshalIndent(res, "  ", "  ")
	if e != nil {
		t.Error(e)
		return
	}
	t.Logf("After add new version\n %v", string(j))
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
	e = DeleteVersion(res.Id, &res.Versions[0])
	if e != nil {
		t.Error(e)
	}
}
