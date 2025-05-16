package configs

import (
	"net/url"
	"strings"
)



type Backends struct {
    envURLs string
}

func NewBackends(envURLs string) *Backends {
    return &Backends{envURLs: envURLs}
}

func (r *Backends) GetBackends() ([]*url.URL, error) {
    rawUrls := strings.Split(r.envURLs, ",")
    var urls []*url.URL
    
    for _, u := range rawUrls {
        parsed, err := url.Parse(strings.TrimSpace(u))
        if err != nil {
            return nil, err
        }
        urls = append(urls, parsed)
    }
    
    return urls, nil
}