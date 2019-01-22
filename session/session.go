// Session相关
//   变更历史
//     2017-02-14  lixiaoya  新建
package session

import (
	"errors"
	"fmt"
	"github.com/lixy529/gotools/utils"
	"net/http"
	"time"
)

// 保存数据接口
type SessData interface {
	Id() string
	Set(key string, value interface{}) error
	Get(key string) interface{}
	Delete(key string) error // 删除key对应的Session
	Flush() error            // 清空Session Id下的所有数据
	Write() error            // 将内存里的数据写到memcache、redis等，Action调用完执行
	Read() error             // 从memcache、redis等读取数据到内存，GetSessData时调用
}

// Provider 接口
type Provider interface {
	Init(lifeTime int64, providerConfig string) error // 初始化一个Provider，NewManager时会调用
	GetSessData(id string) (SessData, error)          // 返回一个SessData，SessStart时调用
	Destroy(id string) error                          // 销毁Id对应的SessData
	Gc()                                              // 清过期session
}

var providers = make(map[string]Provider)

// Register 注册provider
//   参数
//     name:    适配器名称
//     provide: 适配器对象
//   返回
//
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, ok := providers[name]; ok {
		panic("session: Register called twice for provider " + name)
	}
	providers[name] = provide
}

// Manager Session管理器
type Manager struct {
	provider       Provider
	lifeTime       int64
	providerConfig string
	cookieName     string
}

// NewManager 实例化一个Session管理器对象
//   参数
//     providerName:   名称
//     providerConfig: 配置信息
//     cookieName:     cookie名称
//     lifeTime:       生命周期
//   返回
//     成功时返回Session管理器对象，失败时返回错误
func NewManager(providerName, providerConfig, cookieName string, lifeTime int64) (*Manager, error) {
	provider, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provider %q (forgotten import?)", providerName)
	}

	// 初始化一个Provider
	if lifeTime <= 0 {
		lifeTime = 3600
	}
	err := provider.Init(lifeTime, providerConfig)
	if err != nil {
		return nil, err
	}

	if len(cookieName) == 0 {
		cookieName = "GOSESSIONID"
	}

	if lifeTime <= 0 {
		lifeTime = 3600
	}

	return &Manager{
		provider:       provider,
		lifeTime:       lifeTime,
		providerConfig: providerConfig,
		cookieName:     cookieName,
	}, nil
}

// createSessId 生成一个Session Id
// 从cookie里取Session Id
// 如果cookie不存在，则生成一个Session Id，并保存到cookie里
//   参数
//     w: ResponseWriter对象
//     r: Request对象
//   返回
//     Session Id
func (m *Manager) createSessId(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 生成Session Id
	sessId := utils.Guid()
	if len(sessId) == 0 {
		panic("session: create session id error")
	}

	// 保存到cookie
	doMain := utils.GetTopDomain(r.Host)
	cookie = &http.Cookie{
		Name:  m.cookieName,
		Value: sessId,
		Path:  "/",
		//HttpOnly: true,
		//Secure:   true,
		Domain: doMain,
	}
	http.SetCookie(w, cookie)

	return sessId

}

// SessStart 开始session
// 1.生成sessionId
// 2.返回provider
//   参数
//     w: ResponseWriter对象
//     r: Request对象
//   返回
//     成功时返回Session适配器对象，失败返回错误信息
func (m *Manager) SessStart(w http.ResponseWriter, r *http.Request) (SessData, error) {
	sessId := m.createSessId(w, r)
	if len(sessId) == 0 {
		return nil, errors.New("session: create session id error")
	}

	sessData, err := m.provider.GetSessData(sessId)
	if err != nil {
		return nil, err
	}

	return sessData, nil
}

// GetSessData 获取Session Data
//   参数
//     id: Session Id
//   返回
//     成功时返回Session适配器对象，失败返回错误信息
func (m *Manager) GetSessData(id string) (sessions SessData, err error) {
	sessData, err := m.provider.GetSessData(id)
	return sessData, err
}

// SessDestroy 销毁当前会话的Session
//   参数
//     w: ResponseWriter对象
//     r: Request对象
//   返回
//
func (m *Manager) SessDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	// 清数据
	id := cookie.Value
	if m.provider != nil {
		m.provider.Destroy(id)
	}

	// 清cookie
	doMain := utils.GetTopDomain(r.Host)
	cookie = &http.Cookie{
		Name:    m.cookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Now(),
		MaxAge:  -1,
		//HttpOnly: true,
		//Secure:  true,
		Domain: doMain,
	}

	http.SetCookie(w, cookie)
	return
}

// SessGc 定时删除过期session
//   参数
//
//   返回
//     成功时返回Session适配器对象，失败返回错误信息
func (m *Manager) SessGc() {
	m.provider.Gc()
	time.AfterFunc(time.Duration(m.lifeTime)*time.Second, func() { m.SessGc() })
}
