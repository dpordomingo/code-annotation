package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	netContext "golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// OAuthConfig defines enviroment variables for OAuth
type OAuthConfig struct {
	ClientID     string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret string `envconfig:"CLIENT_SECRET" required:"true"`
}

// OAuth service abstracts OAuth implementation
type OAuth struct {
	config             *oauth2.Config
	store              *sessions.CookieStore
	stateGenerator     StateGenerator
	oauthClientBuilder OAuthClientBuilder
	oauthExchangeFunc  OauthExchangeFunc
}

// StateGenerator is a func that returns a random state
type StateGenerator func() string

// DefaultStateGenerator generates a random string using rand.Rand
func DefaultStateGenerator() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// OAuthClientBuilder is a func that returns an http.Client for certain oauth configuration
type OAuthClientBuilder func(ctx context.Context, token *oauth2.Token, conf *oauth2.Config) Getter

// NewOAuthClient returns a http.Client for the passed oauth conf, token and context
func NewOAuthClient(ctx context.Context, token *oauth2.Token, conf *oauth2.Config) Getter {
	return conf.Client(ctx, token)
}

// Getter modelates a the Get method of an http.Client
type Getter interface {
	Get(url string) (*http.Response, error)
}

// OauthExchangeFunc is a func that converts an authorization code into a token.
type OauthExchangeFunc func(netContext.Context, string, *oauth2.Config) (*oauth2.Token, error)

// DefaultOauthExchangeFunc converts an authorization code into a token using the passed oauth config
func DefaultOauthExchangeFunc(ctx netContext.Context, code string, conf *oauth2.Config) (*oauth2.Token, error) {
	return conf.Exchange(ctx, code)
}

// NewOAuth return new OAuth service
func NewOAuth(
	clientID string,
	clientSecret string,
	generator StateGenerator,
	oauthClientBuilder OAuthClientBuilder,
	oauthExchangeFunc OauthExchangeFunc,
) *OAuth {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"read:user"},
		Endpoint:     github.Endpoint,
	}
	return &OAuth{
		config:             config,
		store:              sessions.NewCookieStore([]byte(clientSecret)),
		stateGenerator:     generator,
		oauthClientBuilder: oauthClientBuilder,
		oauthExchangeFunc:  oauthExchangeFunc,
	}
}

// GithubUser represents the user response returned by the GitHub auth.
type GithubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Username  string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// MakeAuthURL returns string for redirect to provider
func (o *OAuth) MakeAuthURL() (string, string) {
	state := o.stateGenerator()
	return o.config.AuthCodeURL(state), state
}

func (o *OAuth) GetState(r *http.Request) interface{} {
	session, _ := o.store.Get(r, "sess")
	return session.Values["state"]
}

// StoreState stores the passed state into the session
func (o *OAuth) StoreState(w http.ResponseWriter, r *http.Request, state string) error {
	session, _ := o.store.Get(r, "sess")
	session.Values["state"] = state

	return session.Save(r, w)
}

// ValidateState protects the user from CSRF attacks
func (o *OAuth) ValidateState(r *http.Request, state string) error {
	session, err := o.store.Get(r, "sess")
	if err != nil {
		return fmt.Errorf("can't get session: %s", err)
	}
	if state != session.Values["state"] {
		return fmt.Errorf("incorrect state: %s", state)
	}
	return nil
}

// GetUser gets user from provider and return user model
func (o *OAuth) GetUser(ctx context.Context, code string) (*GithubUser, error) {
	token, err := o.oauthExchangeFunc(ctx, code, o.config)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange error: %s", err)
	}
	client := o.oauthClientBuilder(ctx, token, o.config)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("can't get user from github: %s", err)
	}
	defer resp.Body.Close()
	var user GithubUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("can't parse github response: %s", err)
	}

	return &user, nil
}
