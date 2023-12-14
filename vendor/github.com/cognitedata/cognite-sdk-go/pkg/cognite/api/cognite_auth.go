package api

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"net/url"
)

type CogniteAuth interface {
	ConfigureAuth(req *http.Request)
}

type apiKeyAuth struct {
	ApiKey string
}

func NewApiKeyAuth(apiKey string) CogniteAuth {
	authConfig := new(apiKeyAuth)
	authConfig.ApiKey = apiKey
	return authConfig
}

func (auth *apiKeyAuth) ConfigureAuth(req *http.Request) {
	req.Header.Add("api-key", auth.ApiKey)
}

type tokenSourceAuth struct {
	Token func() (string, error)
}

func (auth *tokenSourceAuth) ConfigureAuth(req *http.Request) {
	token, err := auth.Token()
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
}

func NewTokenSourceAuth(tokenSource func() (string, error)) CogniteAuth {
	auth := new(tokenSourceAuth)
	auth.Token = tokenSource
	return auth
}

func NewTokenAuth(token string) CogniteAuth {
	tokenSource := func() (string, error) {
		return token, nil
	}
	return NewTokenSourceAuth(tokenSource)
}

func NewOidcClientCredsAuthWithParams(tokenUrl string, clientId string, clientSecret string, scopes []string, endpointParams url.Values) CogniteAuth {
	config := clientcredentials.Config{
		ClientID:       clientId,
		ClientSecret:   clientSecret,
		TokenURL:       tokenUrl,
		Scopes:         scopes,
		EndpointParams: endpointParams,
	}
	oauthTokenSource := config.TokenSource(context.Background())
	tokenSource := func() (string, error) {
		token, err := oauthTokenSource.Token()
		if err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}
	return NewTokenSourceAuth(tokenSource)
}

func NewOidcClientCredsAuth(tokenUrl string, clientId string, clientSecret string, scopes []string) CogniteAuth {
	return NewOidcClientCredsAuthWithParams(tokenUrl, clientId, clientSecret, scopes, make(map[string][]string))
}
