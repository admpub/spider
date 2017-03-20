// 结果收集与输出
package collector

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/admpub/spider/app/pipeline/collector/data"
	"github.com/admpub/spider/app/spider"
	"github.com/admpub/spider/config"
	"github.com/admpub/spider/logs"
	"github.com/admpub/spider/runtime/cache"
)

// 结果收集与输出
type Collector struct {
	*spider.Spider                    //绑定的采集规则
	DataChan       chan data.DataCell //文本数据收集通道
	FileChan       chan data.FileCell //文件收集通道
	dataDocker     []data.DataCell    //分批输出结果缓存
	outType        string             //输出方式
	fileOutType    string             //文件输出方式
	FileOutPath    string             //文件输出路径
	// size     [2]uint64 //数据总输出流量统计[文本，文件]，文本暂时未统计
	dataBatch   uint64 //当前文本输出批次
	fileBatch   uint64 //当前文件输出批次
	wait        sync.WaitGroup
	sum         [4]uint64 //收集的数据总数[上次输出后文本总数，本次输出后文本总数，上次输出后文件总数，本次输出后文件总数]，非并发安全
	dataSumLock sync.RWMutex
	fileSumLock sync.RWMutex
	logger      logs.Logs
	dockerCap   int
}

func NewCollector(sp *spider.Spider) *Collector {
	var self = &Collector{}
	self.Spider = sp
	if len(sp.OutType) > 0 {
		self.outType = sp.OutType
	} else {
		self.outType = cache.Task.OutType
	}
	if len(sp.FileOutType) > 0 {
		self.fileOutType = sp.FileOutType
	} else {
		self.fileOutType = cache.Task.FileOutType
	}
	if len(self.fileOutType) == 0 {
		self.fileOutType = `local`
	}
	if len(sp.FileOutPath) > 0 {
		self.FileOutPath = sp.FileOutPath
	} else {
		self.FileOutPath = cache.Task.FileOutPath
	}
	if len(self.FileOutPath) == 0 {
		self.FileOutPath = config.FILE_DIR
	}
	if sp.DockerCap > 0 {
		self.dockerCap = sp.DockerCap
	} else {
		if cache.Task.DockerCap < 1 {
			cache.Task.DockerCap = 1
		}
		self.dockerCap = cache.Task.DockerCap
	}
	self.DataChan = make(chan data.DataCell, self.dockerCap)
	self.FileChan = make(chan data.FileCell, self.dockerCap)
	self.dataDocker = make([]data.DataCell, 0, self.dockerCap)
	self.sum = [4]uint64{}
	// self.size = [2]uint64{}
	self.dataBatch = 0
	self.fileBatch = 0
	return self
}

func (self *Collector) CollectData(dataCell data.DataCell) error {
	var err error
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("输出协程已终止")
		}
	}()
	self.DataChan <- dataCell
	return err
}

func (self *Collector) CollectFile(fileCell data.FileCell) error {
	var err error
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("输出协程已终止")
		}
	}()
	self.FileChan <- fileCell
	return err
}

// 停止
func (self *Collector) Stop() {
	go func() {
		defer func() {
			recover()
		}()
		close(self.DataChan)
	}()
	go func() {
		defer func() {
			recover()
		}()
		close(self.FileChan)
	}()
}

// 启动数据收集/输出管道
func (self *Collector) Start() {
	// 启动输出协程
	go self.Run()
}

// 启动数据收集/输出管道
func (self *Collector) Run() {
	dataStop := make(chan bool)
	fileStop := make(chan bool)

	go func() {
		defer func() {
			recover()
		}()
		for data := range self.DataChan {
			// 缓存分批数据
			self.dataDocker = append(self.dataDocker, data)

			// 未达到设定的分批量时继续收集数据
			if len(self.dataDocker) < self.dockerCap {
				continue
			}
			// 执行输出
			self.dataBatch++
			self.outputData()
		}
		// 将剩余收集到但未输出的数据输出
		self.dataBatch++
		self.outputData()
		close(dataStop)
	}()

	go func() {
		defer func() {
			recover()
		}()
		// 只有当收到退出通知并且通道内无数据时，才退出循环
		for file := range self.FileChan {
			atomic.AddUint64(&self.fileBatch, 1)
			self.wait.Add(1)
			go self.outputFile(file)
		}
		close(fileStop)
	}()

	<-dataStop
	<-fileStop

	// 等待所有输出完成
	self.wait.Wait()

	// 返回报告
	self.Report()
}

func (self *Collector) resetDataDocker() {
	for _, cell := range self.dataDocker {
		data.PutDataCell(cell)
	}
	self.dataDocker = self.dataDocker[:0]
}

// 获取文本数据总量
func (self *Collector) dataSum() uint64 {
	self.dataSumLock.RLock()
	defer self.dataSumLock.RUnlock()
	return self.sum[1]
}

// 更新文本数据总量
func (self *Collector) addDataSum(add uint64) {
	self.dataSumLock.Lock()
	defer self.dataSumLock.Unlock()
	self.sum[0] = self.sum[1]
	self.sum[1] += add
}

// 获取文件数据总量
func (self *Collector) fileSum() uint64 {
	self.fileSumLock.RLock()
	defer self.fileSumLock.RUnlock()
	return self.sum[3]
}

// 更新文件数据总量
func (self *Collector) addFileSum(add uint64) {
	self.fileSumLock.Lock()
	defer self.fileSumLock.Unlock()
	self.sum[2] = self.sum[3]
	self.sum[3] += add
}

// // 获取文本输出流量
// func (self *Collector) dataSize() uint64 {
// 	return self.size[0]
// }

// // 更新文本输出流量记录
// func (self *Collector) addDataSize(add uint64) {
// 	self.size[0] += add
// }

// // 获取文件输出流量
// func (self *Collector) fileSize() uint64 {
// 	return self.size[1]
// }

// // 更新文本输出流量记录
// func (self *Collector) addFileSize(add uint64) {
// 	self.size[1] += add
// }

// 返回报告
func (self *Collector) Report() {
	cache.ReportChan <- &cache.Report{
		SpiderName: self.Spider.GetName(),
		Keyin:      self.GetKeyin(),
		DataNum:    self.dataSum(),
		FileNum:    self.fileSum(),
		// DataSize:   self.dataSize(),
		// FileSize: self.fileSize(),
		Time: time.Since(cache.StartTime),
	}
}

func (self *Collector) Logger() logs.Logs {
	if self.logger == nil {
		return logs.Log
	}
	return self.logger
}

func (self *Collector) SetLogger(logger logs.Logs) {
	self.logger = logger
}
