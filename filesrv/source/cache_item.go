package source

type CacheItem struct {
	ItemId     string            `json:"itemId"`
	Hashes     map[string]string `json:"hashes"`
	CachedPath string            `json:"cachedPath"`
}

func (p *CacheItem) IsHashEqual(hashes map[string]string) bool {
	if len(p.Hashes) == 0 || len(hashes) == 0 {
		return false
	}
	for k, v := range p.Hashes {
		if value, ok := hashes[k]; ok && value == v {
			continue
		}
		return false
	}
	return true
}
