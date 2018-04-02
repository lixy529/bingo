// 内存Session
//   变更历史
//     2017-02-14  lixiaoya  新建
package memory

import (
	"github.com/lixy529/bingo/session"
	"sync"
	"time"
)

// MemData Session在内存里保存的数据单元
type MemData struct {
	id      string
	accTime time.Time // 最后一次访问时间
	values  map[string]interface{}
	lock    sync.RWMutex
}

// Id 返回Session Id
//   参数
//
//   返回
//     Session Id
func (d *MemData) Id() string {
	return d.id
}

// Set 根据key获取value
//   参数
//     key:   Session的Key值
//     value: Session的Value值
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemData) Set(key string, value interface{}) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.values == nil {
		d.values = make(map[string]interface{})
	}
	d.values[key] = value

	return nil
}

// Get 根据key获取value
//   参数
//     key: Session的Key值
//   返回
//     Session的Value值
func (d *MemData) Get(key string) interface{} {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.values == nil {
		return nil
	}

	if v, ok := d.values[key]; ok {
		return v
	}
	return nil
}

// Delete 删除key
//   参数
//     key: Session的Key值
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemData) Delete(key string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.values, key)
	return nil
}

// Flush 清楚所有的数据
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemData) Flush() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.values = make(map[string]interface{})
	return nil
}

// Write 将数据写到存储对象，内存Session不需要
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemData) Write() error {
	return nil
}

// Read 从存储对象读取数据到内存，内存Session不需要
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemData) Read() error {
	return nil
}

// MemProvider 继承Provider接口
type MemProvider struct {
	curId     string
	lock      sync.RWMutex
	sessDatas map[string]MemData
	lifeTime  int64
}

// Init 初始化MemProvider
//   参数
//     lifeTime: session超时时间
//     providerConfig: 配置，内存session不需要此项
//   返回
//     成功时返回nil，失败返回错误信息
func (p *MemProvider) Init(lifeTime int64, providerConfig string) error {
	p.curId = ""
	if lifeTime <= 0 {
		p.lifeTime = 3600
	} else {
		p.lifeTime = lifeTime
	}

	return nil
}

// GetSessData 从sessDatas取出Session Id对应的Data
// 如果存在，则返回对应Data
// 如果不存在，则生成一个Data，并保存到sessDatas里
//   参数
//     id: Session Id
//   返回
//     成功时Session适配器对象，失败返回错误信息
func (p *MemProvider) GetSessData(id string) (session.SessData, error) {
	p.curId = id

	sessData := func() session.SessData {
		p.lock.RLock()
		defer p.lock.RUnlock()
		if sessData, ok := p.sessDatas[id]; ok {
			return &sessData
		} else {
			return nil
		}
	}()
	if sessData != nil {
		return sessData, nil
	}

	// 不存在，初始化一个
	sessData = func() session.SessData {
		newData := &MemData{id: id, accTime: time.Now(), values: make(map[string]interface{})}
		p.lock.Lock()
		defer p.lock.Unlock()
		p.sessDatas[id] = *newData

		return newData
	}()

	return sessData, nil
}

// Destroy 销毁Id对应的session
//   参数
//     id: Session Id
//   返回
//     成功时nil，失败返回错误信息
func (p *MemProvider) Destroy(id string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if sessData, ok := p.sessDatas[id]; ok {
		sessData.Flush()
		delete(p.sessDatas, id)
		return nil
	}
	return nil
}

// Gc 清过期session
//   参数
//
//   返回
//
func (p *MemProvider) Gc() {
	p.lock.RLock()
	for key, sessData := range p.sessDatas {
		if time.Now().Unix()-sessData.accTime.Unix() > p.lifeTime {
			p.lock.RUnlock()
			p.lock.Lock()
			sessData.Flush()
			delete(p.sessDatas, key)
			p.lock.Unlock()
			p.lock.RLock()
		}
	}
	p.lock.RUnlock()
}

var memProvider = &MemProvider{curId: "", sessDatas: make(map[string]MemData)}

// init 初始化
//   参数
//
//   返回
//
func init() {
	session.Register("memory", memProvider)
}
