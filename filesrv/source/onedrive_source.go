package source

import (
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"xxtuitui.com/filesvr/msgraphapi"
)

type (
	OneDriveContext struct {
		ClientId        string      `json:"clientId"`
		ClientSecret    string      `json:"clientSecret"`
		User            string      `json:"user"`
		LastRefreshTime time.Time   `json:"lastRefreshTime"`
		TenantId        string      `json:"tenantId"`
		CachedItems     []CacheItem `json:"cachedItems"`
	}

	OneDriveSource struct {
		Context *OneDriveContext
		mapping map[string]*CacheItem
		Client  *msgraphapi.MSGraphClient
	}
)

func (p *OneDriveSource) Restore(context *CacheSourceContext) error {
	var sourceContext OneDriveContext
	mapstructure.Decode(context.Context, &sourceContext)
	context.Context = &sourceContext
	p.Context = &sourceContext

	if len(p.Context.ClientId) == 0 || len(p.Context.ClientSecret) == 0 || len(p.Context.User) == 0 {
		return errors.New("InvalidContext")
	}
	if err := p.Init(
		p.Context.ClientId,
		p.Context.ClientSecret,
		"https://graph.microsoft.com/.default",
		p.Context.TenantId); err != nil {
		return err
	}
	return nil
}

func (p *OneDriveSource) Init(clientId string, clientSecret string, scope string, tenantId string) error {
	if p.mapping == nil {
		p.mapping = make(map[string]*CacheItem)
	}
	p.Client = msgraphapi.NewMSGraphClient("https://graph.microsoft.com/v1.0")
	_, err := p.Client.GetToken(clientId, clientSecret, scope, tenantId)
	if err != nil {
		return err
	}
	user, err := p.Client.GetUser(p.Context.User)
	if err != nil {
		return err
	}
	p.Client.SetDefaultUserId(user.Id)
	p.Context.LastRefreshTime = time.Now()
	logrus.WithFields(logrus.Fields{
		"accessToken": p.Client.Token,
	}).Info("OneDriveSourceInitialized")
	return nil
}

func (p *OneDriveSource) MappingFile(reqFileUrl string, localName string, hashes map[string]string) error {
	for _, item := range p.Context.CachedItems {
		if !item.IsHashEqual(hashes) {
			continue
		}
		p.mapping[reqFileUrl] = &item
		logrus.WithFields(logrus.Fields{
			"reqUrl": item.ItemId,
		}).Info("MappingFile")
		return nil
	}
	return errors.New("CachedFileNotFound")
}

func (p *OneDriveSource) GetUrl(reqFileUrl string) (string, error) {
	var item *CacheItem
	if value, ok := p.mapping[reqFileUrl]; ok {
		item = value
	} else {
		return "", errors.New("MappingFileNotFound")
	}
	f, err := p.Client.GetDriveItemById(item.ItemId)
	if err != nil {
		return "", err
	}
	return f.DownloadUrl, nil
}

func (p *OneDriveSource) RefreshSource() ([]CacheItem, error) {
	files, err := p.Client.ListFileRecursiveByPath("/")
	if err != nil {
		return nil, err
	}
	res := []CacheItem{}
	for _, item := range files {
		res = append(res, CacheItem{
			ItemId:     item.Id,
			Hashes:     map[string]string{"quickxorhash": item.File.Hashes.QuickXorHash},
			CachedPath: item.ParentReference.Path,
		})
	}
	logrus.WithFields(logrus.Fields{
		"count": len(res),
	}).Info("OneDriveRefreshSource")
	p.Context.CachedItems = res
	return res, nil
}

func (p *OneDriveSource) RestoreSource(items *[]CacheItem) {
	p.Context.CachedItems = *items
}

func (p *OneDriveSource) MappedFileSize() int { return len(p.mapping) }

func (p *OneDriveSource) CachedFileSize() int { return len(p.Context.CachedItems) }

func (p *OneDriveSource) HasMapping(reqUrl string) bool {
	_, ok := p.mapping[reqUrl]
	return ok
}
