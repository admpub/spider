package collector

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/admpub/spider/app/pipeline/collector/data"
	"github.com/admpub/spider/common/util"
)

func init() {
	FileOutput["local"] = func(self *Collector, file data.FileCell) (fileName string, size int64, err error) {
		// 路径： file/"RuleName"/"time"/"Name"
		p, n := filepath.Split(filepath.Clean(file["Name"].(string)))
		// dir := filepath.Join(self.FileOutPath, util.FileNameReplace(self.namespace())+"__"+cache.StartTime.Format("2006年01月02日 15时04分05秒"), p)
		dir := filepath.Join(self.FileOutPath, util.FileNameReplace(self.namespace()), p)
		// 文件名
		fileName = filepath.Join(dir, util.FileNameReplace(n))
		var d os.FileInfo
		// 创建/打开目录
		d, err = os.Stat(dir)
		if err != nil || !d.IsDir() {
			if err = os.MkdirAll(dir, 0777); err != nil {
				return
			}
		}
		var f *os.File
		// 文件不存在就以0777的权限创建文件，如果存在就在写入之前清空内容
		f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return
		}
		defer f.Close()

		size, err = io.Copy(f, bytes.NewReader(file["Bytes"].([]byte)))
		return
	}
}
