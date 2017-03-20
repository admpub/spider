package collector

import (
	"sync/atomic"

	"github.com/admpub/spider/app/pipeline/collector/data"
	bytesSize "github.com/admpub/spider/common/bytes"
	// "github.com/admpub/spider/runtime/cache"
)

// 文件输出
func (self *Collector) outputFile(file data.FileCell) {
	// 复用FileCell
	defer func() {
		data.PutFileCell(file)
		self.wait.Done()
	}()

	output, ok := FileOutput[self.fileOutType]
	if !ok {
		self.Logger().Error(`Invaid fileOutType: %v`, self.fileOutType)
		return
	}
	// 执行输出
	fileName, size, err := output(self, file)
	if err != nil {
		self.Logger().Error(" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v [ERROR]  %v\n",
			self.Spider.GetName(), self.Spider.GetKeyin(), atomic.LoadUint64(&self.fileBatch), fileName, err)
		return
	}
	// 输出统计
	self.addFileSum(1)

	// 打印报告
	self.Logger().Informational(" * ")
	self.Logger().App(
		" *     [文件下载：%v | KEYIN：%v | 批次：%v]   %v (%s)\n",
		self.Spider.GetName(), self.Spider.GetKeyin(), atomic.LoadUint64(&self.fileBatch), fileName, bytesSize.Format(uint64(size)),
	)
	self.Logger().Informational(" * ")
}
