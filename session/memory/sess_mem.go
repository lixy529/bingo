// Memory Session.
package memory

import (
	"github.com/lixy529/bingo/session"
	"sync"
	"time"
)

// MemData session's data unit in memory.
type MemData struct {
	id      string
	accTime time.Time // Last visit time
	values  map[string]interface{}
	lock    sync.RWMutex
}

// Id return Session Id.
func (d *MemData) Id() string {
	return d.id
}

// Set set value by key.
func (d *MemData) Set(key string, value interface{}) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.values == nil {
		d.values = make(map[string]interface{})
	}
	d.values[key] = value

	return nil
}

// Get return value by key.
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

// Delete delete value by key.
func (d *MemData) Delete(key string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.values, key)
	return nil
}

// Flush clear session by Id.
func (d *MemData) Flush() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.values = make(map[string]interface{})
	return nil
}

// Write write memory data to storage object, memory Session is not required.
func (d *MemData) Write() error {
	return nil
}

// Read read data from storage object, memory Session is not required.
func (d *MemData) Read() error {
	return nil
}

// MemProvider inheriting Provider interface.
type MemProvider struct {
	curId     string
	lock      sync.RWMutex
	sessDatas map[string]MemData
	lifeTime  int64
}

// Init initialize MemProvider.
// lifeTime: Session timeout.
// providerConfig: Session config, memory session is not required.
func (p *MemProvider) Init(lifeTime int64, providerConfig string) error {
	p.curId = ""
	if lifeTime <= 0 {
		p.lifeTime = 3600
	} else {
		p.lifeTime = lifeTime
	}

	return nil
}

// GetSessData return a SessData by ID.
// If the ID isn't exist, create a ID and saved in p.sessDatas.
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

	// not exist, init.
	sessData = func() session.SessData {
		newData := &MemData{id: id, accTime: time.Now(), values: make(map[string]interface{})}
		p.lock.Lock()
		defer p.lock.Unlock()
		p.sessDatas[id] = *newData

		return newData
	}()

	return sessData, nil
}

// Destroy destroy SessData by Id.
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

// Gc clear expired sessions.
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

// init register a memory session provider.
func init() {
	session.Register("memory", memProvider)
}
