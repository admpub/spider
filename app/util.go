package app

import (
	"io"

	"github.com/admpub/spider/app/crawler"
	"github.com/admpub/spider/app/spider"
	"github.com/admpub/spider/logs"
	"github.com/admpub/spider/runtime/cache"
)

func RunCrawler(name string, logWriter io.Writer, mws ...func(*spider.Spider)) *cache.Report {
	logger := logs.NewLog()
	logger.BeeLogger.Async(false)
	logger.SetOutput(logWriter)

	_spider := LogicApp.GetSpiderByName(name)
	if _spider == nil {
		logger.Error(`没有找到规则：` + name)
		return nil
	}

	_spider.SuccessInherit = false
	_spider.FailureInherit = false

	c := crawler.New(0)
	if len(mws) > 0 {
		mws[0](_spider)
	}
	c.Init(_spider, logger).Run()
	return c.Report(true)
}
