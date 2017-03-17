package crawler

import (
	"github.com/admpub/spider/app/spider"
	"github.com/admpub/spider/common/util"
	"github.com/admpub/spider/logs"
)

// 采集引擎中规则队列
type (
	SpiderQueue interface {
		Reset() //重置清空队列
		Add(*spider.Spider)
		AddAll([]*spider.Spider)
		AddKeyins(string) //为队列成员遍历添加Keyin属性，但前提必须是队列成员未被添加过keyin
		GetByIndex(int) *spider.Spider
		GetByName(string) *spider.Spider
		GetAll() []*spider.Spider
		Len() int // 返回队列长度
	}
	sq struct {
		list []*spider.Spider
	}
)

func NewSpiderQueue() SpiderQueue {
	return &sq{
		list: []*spider.Spider{},
	}
}

func (self *sq) Reset() {
	self.list = []*spider.Spider{}
}

func (self *sq) Add(sp *spider.Spider) {
	sp.SetId(self.Len())
	self.list = append(self.list, sp)
}

func (self *sq) AddAll(list []*spider.Spider) {
	for _, v := range list {
		self.Add(v)
	}
}

// 添加keyin，遍历蜘蛛队列得到新的队列（已被显式赋值过的spider将不再重新分配Keyin）
func (self *sq) AddKeyins(keyins string) {
	keyinSlice := util.KeyinsParse(keyins)
	if len(keyinSlice) == 0 {
		return
	}

	unit1 := []*spider.Spider{} // 不可被添加自定义配置的蜘蛛
	unit2 := []*spider.Spider{} // 可被添加自定义配置的蜘蛛
	for _, v := range self.GetAll() {
		if v.GetKeyin() == spider.KEYIN {
			unit2 = append(unit2, v)
			continue
		}
		unit1 = append(unit1, v)
	}

	if len(unit2) == 0 {
		logs.Log.Warning("本批任务无需填写自定义配置！\n")
		return
	}

	self.Reset()

	for _, keyin := range keyinSlice {
		for _, v := range unit2 {
			v.Keyins = keyin
			nv := *v
			self.Add((&nv).Copy())
		}
	}
	if self.Len() == 0 {
		self.AddAll(append(unit1, unit2...))
	}

	self.AddAll(unit1)
}

func (self *sq) GetByIndex(idx int) *spider.Spider {
	return self.list[idx]
}

func (self *sq) GetByName(n string) *spider.Spider {
	for _, sp := range self.list {
		if sp.GetName() == n {
			return sp
		}
	}
	return nil
}

func (self *sq) GetAll() []*spider.Spider {
	return self.list
}

func (self *sq) Len() int {
	return len(self.list)
}
