package collector

import (
	"encoding/json"
)

/************************ 测试采集规则时 输出 ***************************/

func init() {
	DataOutput["testing"] = func(self *Collector) error {
		for _, datacell := range self.dataDocker {
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
			b, _ := json.MarshalIndent(datacell, ``, ` `)
			self.Logger().App(`<pre>` + string(b) + `</pre>`)
		}
		return nil
	}
}
