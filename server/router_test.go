package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/src-d/code-annotation/server/dbutil"
	"github.com/src-d/code-annotation/server/model"
	"github.com/src-d/code-annotation/server/repository"
	"github.com/src-d/code-annotation/server/service"

	testingTools "github.com/src-d/code-annotation/server/testing-tools"
	"github.com/stretchr/testify/suite"
	netContext "golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

type RouterTestSuite struct {
	testingTools.RouterTestSuite
	router             http.Handler
	server             *httptest.Server
	oauth              *service.OAuth
	db                 *dbutil.DB
	dbPath             string
	jwtService         *service.JWT
	oauthRedirHostName string
}

func sqliteDSN(path string) string {
	return fmt.Sprintf("sqlite://%s", path)
}

var validGhUser = model.User{
	Login:     "valid-github-user",
	Username:  "Valid User Name",
	AvatarURL: "https://avatars2.githubusercontent.com/u/1234567",
	Role:      model.Requester,
}

func (s *RouterTestSuite) createDB(path string) {
	db, err := dbutil.Open(path, false)
	if err != nil {
		// TODO:
		log.Fatal(err)
	}

	defer db.Close()

	// TODO:
	createUserTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER, login TEXT UNIQUE, username TEXT, avatar_url TEXT, role TEXT,
		PRIMARY KEY (id))`
	if _, err := db.Exec(createUserTableSQL); err != nil {
		return
	}

	// TODO:
	userRepo := &repository.Users{DB: db.SQLDB()}
	s.NoError(userRepo.Create(&validGhUser))

	return
}

type mockHTTPClient struct {
	*http.Client
	willSucceed bool
}

func (c *mockHTTPClient) Get(url string) (*http.Response, error) {
	if !c.willSucceed {
		emptyResponse := &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			StatusCode: http.StatusNotFound,
		}

		return emptyResponse, nil
	}

	validGhUser := service.GithubUser{
		ID:        1234567,
		Login:     validGhUser.Login,
		Username:  validGhUser.Username,
		AvatarURL: validGhUser.AvatarURL,
	}

	resp := httptest.NewRecorder()
	body, _ := json.Marshal(validGhUser)
	resp.Write(body)

	return resp.Result(), nil
}

// NewOAuthClient returns a mockHTTPClient
func mockedOAuthClient(ctx context.Context, token *oauth2.Token, conf *oauth2.Config) service.Getter {
	var willSucceed bool
	if code, ok := token.Extra("code").(string); ok && code != "wrong" {
		willSucceed = true
	}

	return &mockHTTPClient{willSucceed: willSucceed}
}

func mockedOAuthExchange(ctx netContext.Context, code string, conf *oauth2.Config) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	return token.WithExtra(map[string]interface{}{"code": code}), nil
}

func (s *RouterTestSuite) SetupSuite() {
	s.dbPath = "/projects/src/github.com/src-d/code-annotation/.tmp.testing.db"
	s.oauthRedirHostName = "http://oauth-redirect-url"
	s.oauth = service.NewOAuth("client-id", "client-secret", func() string { return "state" }, mockedOAuthClient, mockedOAuthExchange)
	dsnPath := sqliteDSN(s.dbPath)
	s.createDB(dsnPath)
	db, err := dbutil.Open(dsnPath, true)
	s.NoError(err)

	s.db = &db

	s.jwtService = service.NewJWT("sign-key")

	s.router = Router(
		testingTools.NewDummyTestLogger(),
		s.jwtService,
		s.oauth,
		s.oauthRedirHostName,
		db.SQLDB(),
		"testdata/build",
	)
}

func (s *RouterTestSuite) SetupTest() {
	s.server = httptest.NewServer(s.router)
}

func (s *RouterTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *RouterTestSuite) TearDownSuite() {
	s.db.Close()
	s.NoError(os.Remove(s.dbPath))
}

func (s *RouterTestSuite) GetResponse(method string, path string, body io.Reader, header *http.Header) *http.Response {
	response, err := testingTools.GetResponse(method, s.server.URL+path, body, header)
	if err != nil {
		s.Fail(err.Error())
	}

	return response
}

func (s *RouterTestSuite) TestStatics() {
	indexContent := "index\n"
	whateverContent := "content\n"
	deepContent := "deep content\n"
	staticContent := "static content\n"

	var response *http.Response
	response = s.GetResponse("GET", "/", nil, nil)
	s.AssertResponseBodyStatus(response, 200, indexContent, "index should be served")
	response = s.GetResponse("GET", "/whatever-content.xml", nil, nil)
	s.AssertResponseBodyStatus(response, 200, whateverContent, "whatever-content should be served")
	response = s.GetResponse("GET", "/one-level/deep-content.ext", nil, nil)
	s.AssertResponseBodyStatus(response, 200, deepContent, "deep-content should be served")
	response = s.GetResponse("GET", "/non-folder/does-not-exist.ext", nil, nil)
	s.AssertResponseBodyStatus(response, 200, indexContent, "when content does not exist, index should be served")
	response = s.GetResponse("GET", "/does-not-exist.ext", nil, nil)
	s.AssertResponseBodyStatus(response, 200, indexContent, "when content does not exist, index should be served")
	response = s.GetResponse("GET", "/static/static-thing.css", nil, nil)
	s.AssertResponseBodyStatus(response, 200, staticContent, "existent static content should be delivered as it is")
	response = s.GetResponse("GET", "/static/non-existing-static-thing.css", nil, nil)
	s.AssertResponseStatus(response, http.StatusNotFound, "non existent stuff under static directory should return StatusNotFound")
}

func (s *RouterTestSuite) TestLogin() {
	var response *http.Response
	response = s.GetResponse("GET", "/login", nil, nil)
	redirectURL, _ := s.oauth.MakeAuthURL()
	s.AssertRedirect(response, http.StatusTemporaryRedirect, redirectURL, "redirect to GH should be performed")
}

func (s *RouterTestSuite) TestOauthCallback() {
	var response *http.Response
	response = s.GetResponse("GET", "/oauth-callback?state=invalid", nil, nil)
	s.AssertResponseStatus(response, http.StatusPreconditionFailed, "wrong state in callback should return http internal error")

	response = s.GetResponse("GET", "/oauth-callback?state=state", nil, nil)
	s.AssertResponseStatus(response, http.StatusPreconditionFailed, "require oauth-callback without being redirected from login should fail")

	response = s.GetResponse("GET", "/login", nil, nil)
	header := &http.Header{}
	header.Set("Cookie", response.Header.Get("Set-Cookie"))
	response = s.GetResponse("GET", "/oauth-callback?state=state&code=wrong", nil, header)
	s.AssertResponseStatus(response, http.StatusInternalServerError, "require oauth-callback with a wrong github user will return Internal Server Error")

	response = s.GetResponse("GET", "/login", nil, nil)
	header = &http.Header{}
	header.Set("Cookie", response.Header.Get("Set-Cookie"))
	response = s.GetResponse("GET", "/oauth-callback?state=state&code=valid", nil, header)
	userRepo := &repository.Users{DB: s.db.SQLDB()}
	validStoredUser, err := userRepo.Get(validGhUser.Login)
	s.NoError(err, "should be an user in testing database")
	jwt, err := s.jwtService.MakeToken(&model.User{ID: validStoredUser.ID})
	s.NoError(err, "jwt creation should not fail")
	redirectURL := fmt.Sprintf("%s/?token=%s", s.oauthRedirHostName, jwt)
	s.AssertRedirect(response, http.StatusTemporaryRedirect, redirectURL, "oauth-callback with a valid github user will redirect to index")
}

func (s *RouterTestSuite) TestUserData() {

	// TODO: no userID in context -> 500

	// TODO: userID in context is not in DB -> 404

	// TODO: success
}

func (s *RouterTestSuite) TestExperimentsData() {
	// TODO: wrongID format -> 500

	// TODO: no experiment in DB -> 404

	// TODO: success
}

func (s *RouterTestSuite) TestFilePairData() {
	// TODO: wrongID format -> 500

	// TODO: no filepair in DB -> 404

	// TODO: success
}

func (s *RouterTestSuite) TestAssignmentsData() {
	// TODO: wrongID format -> 500

	// TODO: no user ID in context -> 500

	// TODO: no yet assignments: try to create
	// - can not create -> 500
	// - continue

	// TODO: success

}

func (s *RouterTestSuite) TestSaveAnswerData() {
	// TODO: wrongID format -> 500

	// TODO: No user in context -> 500

	// TODO: Can not read -> 500
	// TODO: Can not unmarshall -> 500

	// TODO: Error saving -> 500

	// TODO: success
}
