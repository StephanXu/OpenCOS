package msgraphapi

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type MSGraphGetTokenRsp struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type MSGraphClient struct {
	Token         string
	HttpClient    resty.Client
	BaseUrl       string
	DefaultUserId string
}

func NewMSGraphClient(baseUrl string) *MSGraphClient {
	client := &MSGraphClient{}
	client.BaseUrl = baseUrl
	client.HttpClient = *resty.New()
	client.HttpClient.SetBaseURL(baseUrl)
	return client
}

func (p *MSGraphClient) GetToken(clientId string, clientSecret string, scope string, tenantId string) (string, error) {
	resp, err := p.HttpClient.
		R().
		EnableTrace().
		SetResult(&MSGraphGetTokenRsp{}).
		SetFormData(map[string]string{
			"client_id":     clientId,
			"client_secret": clientSecret,
			"scope":         scope,
			"grant_type":    "client_credentials",
		}).
		Post(fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantId))
	if err != nil {
		return "", err
	}
	p.Token = resp.Result().(*MSGraphGetTokenRsp).AccessToken
	p.HttpClient.SetAuthToken(p.Token)
	return p.Token, nil
}
