package objectstream

import (
	"fmt"
	"goodfs/apiserver/config"
	"goodfs/apiserver/model/meta"
	"goodfs/util"
	"log"
	"net/http"
	"os"
)

//send delete message to object server
//if delete fail write to logs
//TODO send by rabbitmq. delete all backup
func DeleteObject(name string, ver *meta.MetaVersion) {
	ext, _ := util.GetFileExt(name, true)
	req, _ := http.NewRequest(http.MethodDelete, "http://"+ver.Locate+"/objects/"+ver.Hash+ext, nil)
	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != http.StatusOK {
		f, ef := os.OpenFile(config.LogDir + "/delete_object_undo.log", os.O_WRONLY|os.O_CREATE, os.ModeAppend)
		if ef != nil {
			log.Panic(ef)
		}
		defer f.Close()
		_, ef = f.WriteString(fmt.Sprintf("delete %v from %v\n", ver.Hash+ext, ver.Locate))
		if ef != nil {
			log.Panic(ef)
		}
	}
	return
}
