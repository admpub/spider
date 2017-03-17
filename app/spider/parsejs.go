package spider

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/admpub/confl"
	"github.com/admpub/spider/config"
	"github.com/admpub/spider/logs"
	"github.com/robertkrimen/otto"
)

// 蜘蛛规则解释器模型
type (
	SpiderModel struct {
		Name            string      `xml:"Name"`
		Description     string      `xml:"Description"`
		Pausetime       int64       `xml:"Pausetime"`
		EnableLimit     bool        `xml:"EnableLimit"`
		EnableKeyin     bool        `xml:"EnableKeyin"`
		EnableCookie    bool        `xml:"EnableCookie"`
		NotDefaultField bool        `xml:"NotDefaultField"`
		Namespace       *Function   `xml:"Namespace>Script"`
		SubNamespace    *Function   `xml:"SubNamespace>Script"`
		Root            *Function   `xml:"Root>Script"`
		Trunk           []RuleModel `xml:"Rule" json:"Rule"`
	}
	RuleModel struct {
		Name      string    `xml:"name,attr"`
		ParseFunc *Function `xml:"ParseFunc>Script"`
		AidFunc   *Function `xml:"AidFunc>Script"`
	}
	Function struct {
		Script string `xml:",chardata"`
		Param  string `xml:"param,attr"`
		params []string
	}
	FuncParam struct {
		Name string
		Data interface{}
	}
)

func (f *Function) IsEmpty() bool {
	return len(strings.TrimSpace(f.Script)) == 0
}

func (f *Function) SetParams(vm *otto.Otto, params ...*FuncParam) {
	end := len(params) - 1
	if f.params == nil {
		for _, p := range strings.Split(f.Param, `,`) {
			p = strings.TrimSpace(p)
			if len(p) == 0 {
				continue
			}
			f.params = append(f.params, p)
		}
	}
	for idx, p := range f.params {
		if idx > end {
			break
		}
		vm.Set(p, params[idx].Data)
	}
	for idx := len(f.params); idx <= end; idx++ {
		vm.Set(params[idx].Name, params[idx].Data)
	}
}

func init() {
	for _, _m := range getSpiderModels() {
		m := _m //保证闭包变量
		var sp = &Spider{
			Name:            m.Name,
			Description:     m.Description,
			EnableCookie:    m.EnableCookie,
			NotDefaultField: m.NotDefaultField,
			RuleTree:        &RuleTree{Trunk: map[string]*Rule{}},
		}
		sp.Pausetime = m.Pausetime
		if m.EnableLimit {
			sp.Limit = LIMIT
		}
		if m.EnableKeyin {
			sp.Keyins = KEYIN
		}

		if m.Namespace != nil && !m.Namespace.IsEmpty() {
			sp.Namespace = func(self *Spider) string {
				vm := otto.New()
				m.Namespace.SetParams(vm, &FuncParam{Name: `self`, Data: self})
				val, err := vm.Eval(m.Namespace.Script)
				if err != nil {
					logs.Log.Error(" *     动态规则  [Namespace]: %v\n", err)
				}
				s, _ := val.ToString()
				return s
			}
		}

		if m.SubNamespace != nil && !m.SubNamespace.IsEmpty() {
			sp.SubNamespace = func(self *Spider, dataCell map[string]interface{}) string {
				vm := otto.New()
				m.SubNamespace.SetParams(vm, &FuncParam{Name: `self`, Data: self}, &FuncParam{Name: `dataCell`, Data: dataCell})
				val, err := vm.Eval(m.SubNamespace.Script)
				if err != nil {
					logs.Log.Error(" *     动态规则  [SubNamespace]: %v\n", err)
				}
				s, _ := val.ToString()
				return s
			}
		}

		sp.RuleTree.Root = func(ctx *Context) {
			vm := otto.New()
			m.Root.SetParams(vm, &FuncParam{Name: `ctx`, Data: ctx})
			_, err := vm.Eval(m.Root.Script)
			if err != nil {
				logs.Log.Error(" *     动态规则  [Root]: %v\n", err)
			}
		}

		for _, rule := range m.Trunk {
			r := new(Rule)
			r.ParseFunc = func(fn *Function) func(*Context) {
				return func(ctx *Context) {
					vm := otto.New()
					fn.SetParams(vm, &FuncParam{Name: `ctx`, Data: ctx})
					_, err := vm.Eval(fn.Script)
					if err != nil {
						logs.Log.Error(" *     动态规则  [ParseFunc]: %v\n", err)
					}
				}
			}(rule.ParseFunc)

			r.AidFunc = func(fn *Function) func(*Context, map[string]interface{}) interface{} {
				return func(ctx *Context, aid map[string]interface{}) interface{} {
					vm := otto.New()
					fn.SetParams(vm, &FuncParam{Name: `ctx`, Data: ctx}, &FuncParam{Name: `aid`, Data: aid})
					val, err := vm.Eval(fn.Script)
					if err != nil {
						logs.Log.Error(" *     动态规则  [AidFunc]: %v\n", err)
					}
					return val
				}
			}(rule.ParseFunc)
			sp.RuleTree.Trunk[rule.Name] = r
		}
		sp.Register()
	}
}

func getSpiderModels() (ms []*SpiderModel) {
	var typeName = `HTML`
	defer func() {
		if p := recover(); p != nil {
			log.Printf("[E] %s动态规则解析: %v\n", typeName, p)
		}
	}()
	files, err := filepath.Glob(filepath.Join(config.SPIDER_DIR, "*"+config.SPIDER_XML_EXT))
	if err != nil {
		log.Printf("[E] %v\n", err)
	}
	for _, filename := range files {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("[E] %s动态规则[%s]: %v\n", typeName, filename, err)
			continue
		}
		var m SpiderModel
		err = xml.Unmarshal(b, &m)
		if err != nil {
			log.Printf("[E] %s动态规则[%s]: %v\n", typeName, filename, err)
			continue
		}
		ms = append(ms, &m)
	}
	typeName = `YAML`
	files, err = filepath.Glob(filepath.Join(config.SPIDER_DIR, "*"+config.SPIDER_YML_EXT))
	if err != nil {
		log.Printf("[E] %v\n", err)
	}
	for _, filename := range files {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("[E] %s动态规则[%s]: %v\n", typeName, filename, err)
			continue
		}
		var m SpiderModel
		err = confl.Unmarshal(b, &m)
		if err != nil {
			log.Printf("[E] %s动态规则[%s]: %v\n", typeName, filename, err)
			continue
		}
		ms = append(ms, &m)
	}
	return
}
