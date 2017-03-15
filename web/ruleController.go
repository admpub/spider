package web

import "net/http"

var ruleController = &RuleController{}

type RuleController struct {
}

func (r *RuleController) Testing(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	name := q.Get(`name`)
	w.Write([]byte(`Hello:` + name))
	w.WriteHeader(200)
}
