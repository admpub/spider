// 数据收集
package pipeline

import (
	"github.com/admpub/spider/app/pipeline/collector"
	"github.com/admpub/spider/app/pipeline/collector/data"
	"github.com/admpub/spider/app/spider"
	"github.com/admpub/spider/logs"
)

// 数据收集/输出管道
type Pipeline interface {
	Run()                            //启动(阻塞)
	Start()                          //启动
	Stop()                           //停止
	CollectData(data.DataCell) error //收集数据单元
	CollectFile(data.FileCell) error //收集文件
	Logger() logs.Logs
	SetLogger(logs.Logs)
}

func New(sp *spider.Spider) Pipeline {
	return collector.NewCollector(sp)
}
