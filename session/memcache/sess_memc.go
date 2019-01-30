// Memcache Session
// When saving objects such as structs, you need to convert the members of the struct into strings,
// such as adding ToString() functions, Take the string from Session and convert it into a structure.
// You can refer to demo.
package memcache

import (
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/lixy529/bingo/session"
	"github.com/lixy529/gotools/utils"
	"strings"
	"sync"
)

var cliMemc *memcache.Client

// MemData session's data unit in memory.
type MemcData struct {
	id       string
	values   map[string]interface{}
	lock     sync.RWMutex
	lifeTime int64
	isUpd    bool
}

// Id return Session Id.
func (d *MemcData) Id() string {
	return d.id
}

// Set set value by key.
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

// Get return value by key.
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

// Delete delete value by key.
func (d *MemcData) Delete(key string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.values, key)
	d.isUpd = true

	return nil
}

// Flush clear session by Id.
func (d *MemcData) Flush() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.values = make(map[string]interface{})
	d.isUpd = true

	return d.Write()
}

// Write write memory data to memcache and clear memory data.
func (d *MemcData) Write() error {
	var err error
	// write to memcache.
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

	// clear memory data.
	d.values = make(map[string]interface{})

	return err
}

// Read read data from storage objectmemcache.
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

// MemProvider inheriting Provider interface.
type MemcProvider struct {
	curId     string
	lock      sync.RWMutex
	sessDatas map[string]MemcData
	lifeTime  int64
}

// Init initialize MemProvider.
// lifeTime: Session timeout.
// providerConfig: Session config, the IP and Port of Memcache server, such as 127.0.0.1:11211,127.0.0.2:11211.
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

// GetSessData return a SessData by ID.
// If the ID isn't exist, create a ID and saved in p.sessDatas.
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

	// not exist, init.
	sessData = func() session.SessData {
		newData := &MemcData{id: id, lifeTime: p.lifeTime, values: make(map[string]interface{}), isUpd: false}
		p.lock.Lock()
		defer p.lock.Unlock()
		p.sessDatas[id] = *newData

		return newData
	}()

	return sessData, nil
}

// Destroy destroy SessData by Id.
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

// Gc clear expired sessions, memcache clears expired data by itself.
func (p *MemcProvider) Gc() {
	return
}

var memcProvider = &MemcProvider{curId: "", sessDatas: make(map[string]MemcData)}

// init register a memcache session provider.
func init() {
	session.Register("memcache", memcProvider)
}
