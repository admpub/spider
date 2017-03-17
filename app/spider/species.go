package spider

import (
	"fmt"

	"github.com/admpub/spider/common/pinyin"
)

// 蜘蛛种类列表
type SpiderSpecies struct {
	list   []*Spider
	hash   map[string]*Spider
	sorted bool
}

// 全局蜘蛛种类实例
var Species = &SpiderSpecies{
	list: []*Spider{},
	hash: map[string]*Spider{},
}

// 向蜘蛛种类清单添加新种类
func (self *SpiderSpecies) Add(sp *Spider) *Spider {
	name := sp.Name
	for i := 2; true; i++ {
		if _, ok := self.hash[name]; !ok {
			sp.Name = name
			self.hash[sp.Name] = sp
			break
		}
		name = fmt.Sprintf("%s(%d)", sp.Name, i)
	}
	sp.Name = name
	self.list = append(self.list, sp)
	return sp
}

func (self *SpiderSpecies) DelByIndex(index int) {
	end := len(self.list)
	if index > end {
		return
	}
	sp := self.list[index]
	if _, y := self.hash[sp.Name]; y {
		delete(self.hash, sp.Name)
	}
	if index == end {
		self.list = self.list[:index]
	} else {
		self.list = append(self.list[:index], self.list[index+1:]...)
	}
}

func (self *SpiderSpecies) DelByName(name string) {
	_, y := self.hash[name]
	if !y {
		return
	}
	delete(self.hash, name)
	for index, sp := range self.list {
		if sp.Name != name {
			continue
		}
		end := len(self.list) - 1
		if index == end {
			self.list = self.list[:index]
		} else {
			self.list = append(self.list[:index], self.list[index+1:]...)
		}
		break
	}
}

// 获取全部蜘蛛种类
func (self *SpiderSpecies) Get() []*Spider {
	if !self.sorted {
		l := len(self.list)
		initials := make([]string, l)
		newlist := map[string]*Spider{}
		for i := 0; i < l; i++ {
			initials[i] = self.list[i].GetName()
			newlist[initials[i]] = self.list[i]
		}
		pinyin.SortInitials(initials)
		for i := 0; i < l; i++ {
			self.list[i] = newlist[initials[i]]
		}
		self.sorted = true
	}
	return self.list
}

func (self *SpiderSpecies) GetByName(name string) *Spider {
	s, _ := self.hash[name]
	return s
}
