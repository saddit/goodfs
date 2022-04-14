package objectstream

import (
	"fmt"
	"goodfs/apiserver/config"
	"log"
	"net/http"
	"os"
)

//DeleteObject send delete message to object server
//if delete fail write to logs
func DeleteObject(locate, fileName string) {
	req, _ := http.NewRequest(http.MethodDelete, "http://"+locate+"/objects/"+fileName, nil)
	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != http.StatusOK {
		f, ef := os.OpenFile(config.LogDir+"/delete_object_undo.log", os.O_WRONLY|os.O_CREATE, os.ModeAppend)
		if ef != nil {
			log.Print(ef)
		}
		defer f.Close()
		_, ef = f.WriteString(fmt.Sprintf("delete %v from %v\n", fileName, locate))
		if ef != nil {
			log.Print(ef)
		}
	}
	return
}
