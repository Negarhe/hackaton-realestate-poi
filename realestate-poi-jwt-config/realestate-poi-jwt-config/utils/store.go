package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/configs"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

const (
	SessionName = "auth_session"
	SessionKey  = "data"
)

type SessionStore struct {
	Store *sessions.CookieStore
}

type OAuthSession struct {
	PostToken  string `json:"post_token"`
	State      string `json:"state"`
	ReturnUrl  string `json:"return_url"`
	IsBuyer    bool   `json:"is_buyer"`
	SessionKey string `json:"session_key"`
}

func NewSessionStore(cfg *configs.SessionConfig) *SessionStore {
	sessionStore := SessionStore{
		Store: sessions.NewCookieStore([]byte(cfg.AuthKey)),
	}
	sessionStore.Store.Options = &sessions.Options{
		MaxAge:   3600,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
	return &sessionStore
}

func (s *SessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return s.Store.Get(r, name)
}

func (s *SessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return s.Store.Save(r, w, session)
}

func (h *SessionStore) GetExistingSession(w http.ResponseWriter, r *http.Request) (*OAuthSession, error) {
	oauthSession, err := h.Get(r, SessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if oauthSession.IsNew {
		return nil, fmt.Errorf("no existing session found")
	}
	data, ok := oauthSession.Values[SessionKey].([]byte)
	if !ok {
		return nil, fmt.Errorf("session exists but no data found")
	}

	session := &OAuthSession{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}
	return session, nil
}

func (h *SessionStore) SaveSession(w http.ResponseWriter, r *http.Request, session *OAuthSession) (*OAuthSession, error) {
	oauthSession, err := h.Get(r, SessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	sessionJson, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	oauthSession.Values[SessionKey] = sessionJson
	err = h.Save(r, w, oauthSession)
	if err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	log.Println("new session saved successfully")
	return session, nil

}

func (h *SessionStore) CreateNewSession(w http.ResponseWriter, r *http.Request, postToken, returnUrl string, isBuyer bool) (*OAuthSession, error) {
	state := uuid.New().String()
	session := &OAuthSession{
		PostToken: postToken,
		ReturnUrl: returnUrl,
		State:     state,
		IsBuyer:   isBuyer,
	}
	return h.SaveSession(w, r, session)
}
