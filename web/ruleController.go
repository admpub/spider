package web

import (
	"net/http"

	"github.com/admpub/spider/app"
	"github.com/admpub/spider/app/spider"
)

var ruleController = &RuleController{}

type RuleController struct {
}

func (r *RuleController) Testing(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	name := q.Get(`name`)
	app.RunCrawler(name, w, func(_spider *spider.Spider) {
		_spider.OutType = `testing`
		_spider.DataLimit = 1
		_spider.DockerCap = 1
		_spider.DisableAsync = true
	})
	w.Write([]byte(`Hello:` + name))
}
