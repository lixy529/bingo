// Memcache Session，支持Session共享
// 保存结构体等对象时，需要将结体构体里的成员转成字符串，比如添加ToString函数，
// 从Session里取出结构体转换的字符串，再转换成结构体，可以有参考demo
//   变更历史
//     2017-02-17  lixiaoya  新建
package memcache

import (
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/lixy529/bingo/session"
	"github.com/lixy529/bingo/utils"
	"strings"
	"sync"
)

var cliMemc *memcache.Client

// MemData Session在内存里保存的数据单元
type MemcData struct {
	id       string
	values   map[string]interface{}
	lock     sync.RWMutex
	lifeTime int64
	isUpd    bool
}

// Id 返回Session Id
//   参数
//
//   返回
//     Session Id
func (d *MemcData) Id() string {
	return d.id
}

// Set 根据key获取value
//   参数
//     key:   Session的Key值
//     value: Session的Value值
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemcData) Set(key string, value interface{}) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.values == nil {
		d.values = make(map[string]interface{})
	}
	d.values[key] = value
	d.isUpd = true

	return nil
}

// Get 根据key获取value
//   参数
//     key: Session的Key值
//   返回
//     Session的Value值
func (d *MemcData) Get(key string) interface{} {
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
func (d *MemcData) Delete(key string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.values, key)
	d.isUpd = true

	return nil
}

// Flush 清楚所有的数据
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemcData) Flush() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.values = make(map[string]interface{})
	d.isUpd = true

	return d.Write()
}

// Write 将数据写到memcache，同时清空内存的数据
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemcData) Write() error {
	var err error
	// 将数据写到memcache
	if d.isUpd && cliMemc != nil {
		if len(d.values) == 0 {
			err = cliMemc.Delete(d.id)
		} else {
			var val []byte
			val, err = utils.GobEncode(d.values)
			if err != nil {
				return err
			}
			item := memcache.Item{Key: d.id, Value: val, Expiration: int32(d.lifeTime)}
			err = cliMemc.Set(&item)
		}
		d.isUpd = false
	}

	// 清空内存的数据
	d.values = make(map[string]interface{})

	return err
}

// Read 从memcache读取数据到内存
//   参数
//
//   返回
//     成功时返回nil，失败返回错误信息
func (d *MemcData) Read() error {
	if cliMemc == nil {
		return errors.New("session: memcache client is nil")
	}

	if d.id == "" {
		return errors.New("session: Session Id is nil")
	}

	item, err := cliMemc.Get(d.id)
	if err != nil && err == memcache.ErrCacheMiss {
		d.values = make(map[string]interface{})
		return nil
	}

	var kv map[string]interface{}
	if len(item.Value) == 0 {
		d.values = make(map[string]interface{})
	} else {
		err = utils.GobDecode(item.Value, &kv)
		if err != nil {
			return err
		}
	}

	return nil
}

// MemProvider 继承Provider接口
type MemcProvider struct {
	curId     string
	lock      sync.RWMutex
	sessDatas map[string]MemcData
	lifeTime  int64
}

// Init 初始化MemProvider
//   参数
//     lifeTime: session超时时间
//     providerConfig: 配置，memcache服务的Ip和Port，如:127.0.0.1:11211,127.0.0.2:11211，多个服务器用,号分隔
//   返回
//     成功时返回nil，失败返回错误信息
func (p *MemcProvider) Init(lifeTime int64, providerConfig string) error {
	p.curId = ""
	if lifeTime <= 0 {
		p.lifeTime = 3600
	} else {
		p.lifeTime = lifeTime
	}

	if cliMemc == nil {
		memcAddr := strings.Split(providerConfig, ",")
		cliMemc = memcache.New(memcAddr...)
		if cliMemc == nil {
			return fmt.Errorf("session: Connect memcache [%s] failed", providerConfig)
		}
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
func (p *MemcProvider) GetSessData(id string) (session.SessData, error) {
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
		if err := sessData.Read(); err != nil {
			return sessData, err
		}

		return sessData, nil
	}

	// 不存在，初始化一个
	sessData = func() session.SessData {
		newData := &MemcData{id: id, lifeTime: p.lifeTime, values: make(map[string]interface{}), isUpd: false}
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
func (p *MemcProvider) Destroy(id string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if sessData, ok := p.sessDatas[id]; ok {
		sessData.Flush()
		delete(p.sessDatas, id)
		return nil
	}
	return nil
}

// Gc 清过期session，memcache自己清到期数据
//   参数
//
//   返回
//
func (p *MemcProvider) Gc() {
	return
}

var memcProvider = &MemcProvider{curId: "", sessDatas: make(map[string]MemcData)}

// init 初始化
//   参数
//
//   返回
//
func init() {
	session.Register("memcache", memcProvider)
}
