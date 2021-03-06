package reqsession

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goatcms/goatcms/cmsapp/services"
	"github.com/goatcms/goatcms/cmsapp/services/requestdep"
	"github.com/goatcms/goatcore/app"
	"github.com/goatcms/goatcore/dependency"
)

// SessionManager provide session manager for reques (current user)
type SessionManager struct {
	deps struct {
		Request         *http.Request           `request:"Request"`
		Response        http.ResponseWriter     `request:"Response"`
		SessionStorage  services.SessionStorage `dependency:"SessionStorageService"`
		SessionCookieID string                  `config:"?session.cookie.id"`
	}
	dataScope app.DataScope
}

// SessionFactory create a session manager instance
func SessionFactory(dp dependency.Provider) (interface{}, error) {
	s := &SessionManager{}
	if err := dp.InjectTo(&s.deps); err != nil {
		return nil, err
	}
	if s.deps.SessionCookieID == "" {
		s.deps.SessionCookieID = services.SessionCookieID
	}
	return requestdep.Session(s), nil
}

// Init build new session if a session doesn't exist
func (s *SessionManager) Scope() (app.DataScope, error) {
	if s.dataScope != nil {
		return s.dataScope, nil
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	if s.dataScope == nil {
		return nil, fmt.Errorf("Init didn't created DataScope")
	}
	return s.dataScope, nil
}

// Init build new session if a session doesn't exist
func (s *SessionManager) Get(name string) (interface{}, error) {
	scope, err := s.Scope()
	if err != nil {
		return nil, err
	}
	return scope.Get(name)
}

// Init build new session if a session doesn't exist
func (s *SessionManager) Set(name string, value interface{}) error {
	scope, err := s.Scope()
	if err != nil {
		return err
	}
	return scope.Set(name, value)
}

// Init build new session if a session doesn't exist
func (s *SessionManager) Keys() ([]string, error) {
	scope, err := s.Scope()
	if err != nil {
		return nil, err
	}
	return scope.Keys()
}

// Init build new session if a session doesn't exist
func (s *SessionManager) Init() error {
	if _, err := s.deps.Request.Cookie(s.deps.SessionCookieID); err != nil {
		if _, _, err = s.createSession(); err != nil {
			return err
		}
	}
	return nil
}

// Create build new session
func (s *SessionManager) createSession() (string, app.DataScope, error) {
	var sessionID string
	var err error
	sessionID, s.dataScope, err = s.deps.SessionStorage.Create()
	if err != nil {
		return "", nil, err
	}
	lifetime, err := s.deps.SessionStorage.SessionLifetime()
	if err != nil {
		return "", nil, err
	}
	expiration := time.Now().Add(time.Duration(lifetime) * time.Hour)
	cookie := http.Cookie{
		Name:     s.deps.SessionCookieID,
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(s.deps.Response, &cookie)
	return sessionID, s.dataScope, nil
}
