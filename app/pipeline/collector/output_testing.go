package collector

import (
	"encoding/json"
	"errors"

	"github.com/admpub/spider/common/util"
)

/************************ 测试采集规则时 输出 ***************************/

func init() {
	DataOutput["testing"] = func(self *Collector) error {
		if self.Spider.Writer == nil {
			return errors.New(`Spider.Writer is nil`)
		}
		var dataMap = make(map[string][]interface{})

		self.Spider.Writer.Write([]byte(`---------------`))
		for _, datacell := range self.dataDocker {
			subNamespace := util.FileNameReplace(self.subNamespace(datacell))

			for k, v := range datacell["Data"].(map[string]interface{}) {
				datacell[k] = v
			}
			delete(datacell, "Data")
			delete(datacell, "RuleName")
			if !self.Spider.OutDefaultField() {
				delete(datacell, "Url")
				delete(datacell, "ParentUrl")
				delete(datacell, "DownloadTime")
			}
			dataMap[subNamespace] = append(dataMap[subNamespace], datacell)
		}
		b, _ := json.MarshalIndent(dataMap, ``, ` `)
		self.Spider.Writer.Write([]byte(`<pre>`))
		self.Spider.Writer.Write(b)
		self.Spider.Writer.Write([]byte(`</pre>`))
		return nil
	}
}
