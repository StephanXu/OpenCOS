package msgraphapi

import "fmt"

type (
	UserResourceSimple struct {
		DisplayName string `json:"displayName"`
		Id          string `json:"id"`
		Mail        string `json:"mail"`
	}
)

func (p *MSGraphClient) GetUser(email string) (*UserResourceSimple, error) {
	resp, err := p.HttpClient.
		R().
		EnableTrace().
		SetResult(&UserResourceSimple{}).
		Get(fmt.Sprintf("/users/%s", email))
	if err != nil {
		return &UserResourceSimple{}, err
	}
	user := resp.Result().(*UserResourceSimple)
	return user, nil
}

func (p *MSGraphClient) SetDefaultUserId(userId string) {
	p.DefaultUserId = userId
}
