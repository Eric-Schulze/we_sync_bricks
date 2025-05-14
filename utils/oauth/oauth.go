package oauth

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dghubble/oauth1"
)

type OAuthClient struct {
	client *http.Client
}

func NewOAuthClient(consumerKey string, consumerSecret string, token string, tokenSecret string) *OAuthClient {
	// TODO: Error handling
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	clientToken := oauth1.NewToken(token, tokenSecret)

	// httpClient will automatically authorize http.Request's
	return &OAuthClient{client: config.Client(oauth1.NoContext, clientToken)}
}

func (oAuth *OAuthClient) Get(path string) ([]byte, error) {
	resp, err := oAuth.client.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", body)
	}

	return io.ReadAll(resp.Body)
}
