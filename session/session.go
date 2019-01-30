package session

import (
	"errors"
	"fmt"
	"github.com/lixy529/gotools/utils"
	"net/http"
	"time"
)

// SessData session interface.
type SessData interface {
	Id() string
	Set(key string, value interface{}) error
	Get(key string) interface{}
	Delete(key string) error // Delete session by key.
	Flush() error            // Clear session by Id.
	Write() error            // Write memory data to storage objec. Action Call Completes Execution.
	Read() error             // Read data from memcache, redis, etc. to memory, call it when GetSessData.
}

// Provider
type Provider interface {
	Init(lifeTime int64, providerConfig string) error // Init Provider, call it when NewManager.
	GetSessData(id string) (SessData, error)          // Return a SessData, call it wehn SessStart.
	Destroy(id string) error                          // Destroy SessData by Id.
	Gc()                                              // Clear expired sessions.
}

var providers = make(map[string]Provider)

// Register register a provider.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, ok := providers[name]; ok {
		panic("session: Register called twice for provider " + name)
	}
	providers[name] = provide
}

// Manager session manager.
type Manager struct {
	provider       Provider
	lifeTime       int64
	providerConfig string
	cookieName     string
}

// NewManager return a session manager object.
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

// createSessId return a Session Id.
// Get Session Id from cookie.
// If the cookie hasn't Session Id, create a Session Id and saved in the cookie.
func (m *Manager) createSessId(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Create Session Id.
	sessId := utils.Guid()
	if len(sessId) == 0 {
		panic("session: create session id error")
	}

	// Saved in the cookie.
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

// SessStart start session.
// 1) create sessionId.
// 2) return provider.
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

// GetSessData return Session  by ID.
func (m *Manager) GetSessData(id string) (sessions SessData, err error) {
	sessData, err := m.provider.GetSessData(id)
	return sessData, err
}

// SessDestroy destroy Session
func (m *Manager) SessDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	// clear data.
	id := cookie.Value
	if m.provider != nil {
		m.provider.Destroy(id)
	}

	// clear cookie.
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

// SessGc delete expired sessions.
func (m *Manager) SessGc() {
	m.provider.Gc()
	time.AfterFunc(time.Duration(m.lifeTime)*time.Second, func() { m.SessGc() })
}
