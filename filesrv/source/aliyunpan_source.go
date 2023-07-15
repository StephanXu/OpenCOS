package source

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/tickstep/aliyunpan-api/aliyunpan"
	"github.com/tickstep/aliyunpan-api/aliyunpan/apierror"
	"xxtuitui.com/filesvr/config"
)

type (
	AliyunpanContext struct {
		RefreshToken    string      `json:"refreshToken"`
		DriveId         string      `json:"driveId"`
		DeviceId        string      `json:"deviceId"`
		LastRefreshTime time.Time   `json:"lastRefreshTime"`
		CachedItems     []CacheItem `json:"cachedItems"`
	}

	AliyunpanSource struct {
		client  *aliyunpan.PanClient
		mapping map[string]*CacheItem
		Context *AliyunpanContext
	}
)

func (p *AliyunpanSource) Restore(context *CacheSourceContext) error {
	var sourceContext AliyunpanContext
	mapstructure.Decode(context.Context, &sourceContext)
	context.Context = &sourceContext
	p.Context = &sourceContext

	if len(p.Context.RefreshToken) == 0 {
		return errors.New("EmptyRefreshToken")
	}
	logrus.WithFields(logrus.Fields{
		"refreshToken": p.Context.RefreshToken,
		"deviceId":     p.Context.DeviceId,
	}).Info("AliyunpanSourceConfigLoaded")
	return p.Init(p.Context.RefreshToken)
}

func (p *AliyunpanSource) Init(refreshToken string) error {
	if p.mapping == nil {
		p.mapping = make(map[string]*CacheItem)
	}
	webToken, err := aliyunpan.GetAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	p.Context.RefreshToken = webToken.RefreshToken
	appConfig := aliyunpan.AppConfig{
		AppId:     "25dzX3vbYqktVxxX",
		DeviceId:  p.Context.DeviceId,
		UserId:    "",
		Nonce:     0,
		PublicKey: "",
	}
	p.client = aliyunpan.NewPanClient(*webToken, aliyunpan.AppLoginToken{}, appConfig, aliyunpan.SessionConfig{
		DeviceName: "Chrome浏览器",
		ModelName:  "Windows网页版",
	})
	user, err := p.client.GetUserInfo()
	if err != nil {
		return err
	}
	p.Context.DriveId = user.FileDriveId
	p.Context.LastRefreshTime = time.Now()
	appConfig.UserId = user.UserId
	p.client.UpdateAppConfig(appConfig)
	logrus.WithFields(logrus.Fields{
		"originRefreshToken": refreshToken,
		"latestRefreshToken": p.Context.RefreshToken,
		"driveId":            p.Context.DriveId,
		"userId":             user.UserId,
	}).Info("AliyunSourceInitialized")
	r, err := p.client.CreateSession(nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("AliyunSourceCreateSessionSuccess")
	}
	if r != nil && !r.Result {
		logrus.Error("AliyunSourceInitializeFailedUnknown")
	}
	return nil
}

func (p *AliyunpanSource) updateToken(refreshToken string) error {
	webToken, err := aliyunpan.GetAccessTokenFromRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	p.Context.RefreshToken = webToken.RefreshToken
	p.client.UpdateToken(*webToken)
	logrus.WithFields(logrus.Fields{
		"originToken":  refreshToken,
		"updatedToken": p.Context.RefreshToken,
	})
	return nil
}

func (p *AliyunpanSource) MappingFile(reqFileUrl string, localName string, hashes map[string]string) error {
	for _, f := range p.Context.CachedItems {
		if !f.IsHashEqual(hashes) {
			continue
		}
		p.mapping[reqFileUrl] = &f
		return nil
	}
	return errors.New("CachedFileNotFound")
}

func (p *AliyunpanSource) GetUrl(reqFileUrl string) (string, error) {
	var item *CacheItem
	if value, ok := p.mapping[reqFileUrl]; ok {
		item = value
	}
	query := aliyunpan.GetFileDownloadUrlParam{
		DriveId:   p.Context.DriveId,
		FileId:    item.ItemId,
		ExpireSec: 3600 * 4,
	}
	res, err := p.client.GetFileDownloadUrl(&query)
	if err != nil {
		if err.ErrCode() == apierror.ApiCodeAccessTokenInvalid {
			logrus.Info("AliyunpanApiTokenExpired")
			if err := p.Init(p.Context.RefreshToken); err == nil {
				config.SaveContext()
				logrus.WithFields(logrus.Fields{
					"reqUrl": reqFileUrl,
				}).Info("AliyunpanRefreshTokenRetry")
				if res, err := p.client.GetFileDownloadUrl(&query); err == nil {
					return res.Url, nil
				}
			}
		}
		logrus.WithFields(logrus.Fields{
			"reqUrl":  reqFileUrl,
			"errCode": err.ErrCode(),
			"err":     err.Error(),
		}).Info("AliyunpanGetUrlFailed")
		return "", err
	}
	return res.Url, nil
}

func (p *AliyunpanSource) RefreshSource() ([]CacheItem, error) {
	nodes := p.client.FilesDirectoriesRecurseList(p.Context.DriveId, "/", nil)
	files := []CacheItem{}
	for _, f := range nodes {
		if !f.IsFile() {
			continue
		}
		hashContent, err := hex.DecodeString(f.ContentHash)
		if err != nil {
			return nil, err
		}
		files = append(files, CacheItem{
			ItemId:     f.FileId,
			Hashes:     map[string]string{f.ContentHashName: base64.StdEncoding.EncodeToString(hashContent)},
			CachedPath: f.Path,
		})
	}
	logrus.WithFields(logrus.Fields{
		"count": len(files),
	}).Info("AliyunpanRefreshSource")
	p.Context.CachedItems = files
	return files, nil
}

func (p *AliyunpanSource) RestoreSource(items *[]CacheItem) {
	p.Context.CachedItems = *items
}

func (p *AliyunpanSource) MappedFileSize() int { return len(p.mapping) }

func (p *AliyunpanSource) CachedFileSize() int { return len(p.Context.CachedItems) }

func (p *AliyunpanSource) HasMapping(reqUrl string) bool {
	_, ok := p.mapping[reqUrl]
	return ok
}
