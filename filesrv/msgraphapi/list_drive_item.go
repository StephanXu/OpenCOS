package msgraphapi

import (
	"fmt"
	"net/url"
)

type (
	Identity struct {
		DisplayName string `json:"displayName"`
		Id          string `json:"id"`
	}
	IdentitySet struct {
		User        *Identity `json:"user"`
		Application *Identity `json:"application"`
		Device      *Identity `json:"device"`
	}

	HashesSet struct {
		Crc32Hash    string `json:"crc32Hash"`
		Sha1Hash     string `json:"sha1Hash"`
		Sha256Hash   string `json:"sha256Hash"`
		QuickXorHash string `json:"quickXorHash"`
	}
	FileResource struct {
		Hashes   HashesSet `json:"hashes"`
		MimeType string    `json:"mimeType"`
	}

	FileSystemInfoFacet struct {
		CreatedDateTime      string `json:"createdDateTime"`
		LastAccessedDateTime string `json:"lastAccessedDateTime"`
		LastModifiedDateTime string `json:"LastModifiedDateTime"`
	}

	FolderResource struct {
		ChildCount int `json:"childCount"`
	}

	VideoResource struct {
		AudioBitsPerSample    int32   `json:"audioBitsPerSample"`
		AudioChannels         int32   `json:"audioChannels"`
		AudioFormat           string  `json:"audioFormat"`
		AudioSamplesPerSecond int32   `json:"audioSamplesPerSecond"`
		Bitrate               int32   `json:"bitrate"`
		Duration              int64   `json:"duration"`
		FourCC                string  `json:"fourCC"`
		FrameRate             float64 `json:"frameRate"`
		Height                int32   `json:"height"`
		Width                 int32   `json:"width"`
	}

	ItemReference struct {
		DriveId       string `json:"driveId"`
		DriveType     string `json:"driveType"`
		Id            string `json:"id"`
		Name          string `json:"name"`
		Path          string `json:"path"`
		ShareId       string `json:"shareId"`
		SharepointIds string `json:"sharepointIds"`
		SiteId        string `json:"siteId"`
	}

	DriveItem struct {
		CreatedBy            IdentitySet          `json:"createdBy"`
		CreatedDateTime      string               `json:"createdDateTime"`
		CTag                 string               `json:"cTag"`
		File                 *FileResource        `json:"file"`
		FileSystemInfo       *FileSystemInfoFacet `json:"fileSystemInfo"`
		Folder               *FolderResource      `json:"folder"`
		Id                   string               `json:"id"`
		LastModifiedBy       IdentitySet          `json:"lastModifiedBy"`
		LastModifiedDateTime string               `json:"lastModifiedDateTime"`
		Name                 string               `json:"name"`
		ParentReference      *ItemReference       `json:"parentReference"`
		Size                 int64                `json:"size"`
		Video                *VideoResource       `json:"video"`
		DownloadUrl          string               `json:"@microsoft.graph.downloadUrl"`
		WebUrl               string               `json:"webUrl"`
	}
)

type (
	listChildRsp struct {
		Context string      `json:"@odata.context"`
		Value   []DriveItem `json:"value"`
	}
)

func (p *MSGraphClient) listChild(isId bool, pathOrId string) ([]DriveItem, error) {
	var query string
	if isId {
		query = fmt.Sprintf("/users/%s/drive/items/%s/children", p.DefaultUserId, pathOrId)
	} else {
		query = fmt.Sprintf("/users/%s/drive/root:/%s:/children", p.DefaultUserId, url.QueryEscape(pathOrId))
	}
	resp, err := p.HttpClient.
		R().
		EnableTrace().
		SetResult(&listChildRsp{}).
		Get(query)
	if err != nil {
		return nil, err
	}
	return resp.Result().(*listChildRsp).Value, nil
}

func (p *MSGraphClient) ListChildByPath(path string) ([]DriveItem, error) {
	return p.listChild(false, path)
}

func (p *MSGraphClient) ListChildByItemId(id string) ([]DriveItem, error) {
	return p.listChild(true, id)
}

func (p *MSGraphClient) listFileRecursive(isId bool, pathOrId string) ([]DriveItem, error) {
	files, err := p.listChild(isId, pathOrId)
	if err != nil {
		return nil, err
	}

	var res []DriveItem
	for _, f := range files {
		if f.Folder != nil {
			fs, err := p.listFileRecursive(true, f.Id)
			if err != nil {
				return nil, err
			}
			res = append(res, fs...)
			continue
		}
		res = append(res, f)
	}
	return res, nil
}

func (p *MSGraphClient) ListFileRecursiveByPath(path string) ([]DriveItem, error) {
	return p.listFileRecursive(false, path)
}

func (p *MSGraphClient) getDriveItem(isId bool, pathOrId string) (*DriveItem, *ApiError) {
	var query string
	if isId {
		query = fmt.Sprintf("/users/%s/drive/items/%s", p.DefaultUserId, pathOrId)
	} else {
		query = fmt.Sprintf("/users/%s/drive/root:/%s", p.DefaultUserId, url.QueryEscape(pathOrId))
	}
	resp, err := p.HttpClient.R().EnableTrace().SetResult(&DriveItem{}).SetError(&ErrorWrapper{}).Get(query)
	if err != nil {
		return nil, p.ParseError(resp.Error().(*ErrorWrapper))
	}
	if resp.IsError() {
		return nil, p.ParseError(resp.Error().(*ErrorWrapper))
	}
	return resp.Result().(*DriveItem), nil
}

func (p *MSGraphClient) GetDriveItemByPath(path string) (*DriveItem, *ApiError) {
	return p.getDriveItem(false, path)
}

func (p *MSGraphClient) GetDriveItemById(id string) (*DriveItem, *ApiError) {
	return p.getDriveItem(true, id)
}
