package site

import (
	"crawler/internal/domain"
	"crawler/proto"
	"fmt"
	"sync"
)

var (
	siteSlice []domain.Site
	siteMap   = map[proto.Site]domain.Site{
		proto.Site_XIAOHONGSHU: &XHS{},
	}
	once sync.Once
)

// Convert  转换proto中的site为domain的site
func Convert(site proto.Site) (domain.Site, error) {
	s, ok := siteMap[site]
	if !ok {
		return nil, fmt.Errorf("不支持该站点:%v", site)
	}
	return s, nil
}

// GetAll 获取全部站点
func GetAll() []domain.Site {
	once.Do(func() {
		siteSlice = make([]domain.Site, 0, len(siteMap))
		for _, site := range siteMap {
			siteSlice = append(siteSlice, site)
		}
	})

	return siteSlice
}
