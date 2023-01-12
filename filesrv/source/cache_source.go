package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type (
	CacheSourceContext struct {
		Name    string      `json:"name"`
		Type    string      `json:"type"`
		Context interface{} `json:"context"`
	}

	CacheSource interface {
		GetUrl(reqFileUrl string) (string, error)
		MappingFile(reqFileUrl string, localName string, hash map[string]string) error
		RefreshSource() ([]CacheItem, error)
		RestoreSource(items *[]CacheItem)
		CachedFileSize() int
		MappedFileSize() int
		Restore(context *CacheSourceContext) error
	}

	SourcesManager struct {
		sources map[string]CacheSource
	}
)

var Manager SourcesManager

func (p *SourcesManager) Restore(context *CacheSourceContext) error {
	if p.HasSource(context.Name) {
		return errors.New("SourceAlreadyExists")
	}
	if context.Type == "Aliyunpan" {
		p.RegisterSource(context.Name, &AliyunpanSource{})
	} else if context.Type == "OneDriveForBusiness" {
		p.RegisterSource(context.Name, &OneDriveSource{})
	} else {
		return errors.New("SourceTypeNotSupported")
	}
	source := *p.GetSource(context.Name)
	if err := source.Restore(context); err != nil {
		return err
	}
	if source.CachedFileSize() == 0 {
		if _, err := source.RefreshSource(); err != nil {
			return err
		}
	}
	return nil
}

func (p *SourcesManager) RefreshSource() {
	res := make(map[string][]CacheItem)
	for k, cs := range p.sources {
		items, err := cs.RefreshSource()
		if err != nil {
			fmt.Printf("err: %v\n", err)
			continue
		}
		res[k] = items
		logrus.New().WithFields(logrus.Fields{
			"sourceName": k,
			"count":      len(items),
		}).Info("SourceRefreshed")
	}

	content, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	ioutil.WriteFile("sources.json", content, 0666)
}

func (p *SourcesManager) RestoreSource(filename string) {
	files := make(map[string][]CacheItem)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: err: %v\n", err)
		return
	}
	if err := json.Unmarshal(content, &files); err != nil {
		fmt.Printf("Error parsing config: err: %v\n", err)
		return
	}

	for k, cs := range p.sources {
		if config, ok := files[k]; ok {
			cs.RestoreSource(&config)
		}
		fmt.Printf("source %s restored from file %s: %d\n", k, filename, cs.CachedFileSize())
	}
}

func (p *SourcesManager) MappingFile(reqFileUrl string, localName string, hashes map[string]string) error {
	for _, cs := range p.sources {
		if err := cs.MappingFile(reqFileUrl, localName, hashes); err != nil {
			return err
		}
	}
	return nil
}

func (p *SourcesManager) RegisterSource(sourceName string, s CacheSource) {
	if p.sources == nil {
		p.sources = make(map[string]CacheSource)
	}
	p.sources[sourceName] = s
}

func (p *SourcesManager) GetSource(sourceName string) *CacheSource {
	if v, ok := p.sources[sourceName]; ok {
		return &v
	}
	return nil
}

func (p *SourcesManager) HasSource(sourceName string) bool {
	_, ok := p.sources[sourceName]
	return ok
}
