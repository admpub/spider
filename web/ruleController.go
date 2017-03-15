package web

import (
	"net/http"

	"github.com/admpub/spider/app"
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
	c := app.LogicApp.CrawlerPool.Use()
	c.Init(_spider).Run()
	w.Write([]byte(`Hello:` + name))
}
