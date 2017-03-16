package web

import (
	"net/http"

	"github.com/admpub/spider/app"
	"github.com/admpub/spider/app/crawler"
	"github.com/admpub/spider/logs"
	"github.com/admpub/spider/runtime/cache"
)

var ruleController = &RuleController{}

type RuleController struct {
}

func (r *RuleController) Testing(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	name := q.Get(`name`)

	_spider := app.LogicApp.GetSpiderByName(name)
	if _spider == nil {
		w.Write([]byte(`没有找到规则：` + name))
		return
	}

	cache.Task.SuccessInherit = false
	cache.Task.FailureInherit = false

	c := crawler.New(0)
	_spider.OutType = `testing`
	_spider.Writer = w
	_spider.Limit = 1
	logger := logs.NewLog()
	logger.BeeLogger.Async(false)
	logger.SetOutput(Lsc)
	c.Init(_spider, logger).Run()
	w.Write([]byte(`Hello:` + name))
}
