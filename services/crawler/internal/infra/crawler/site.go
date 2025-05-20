package crawler

import (
	"crawler/internal/domain"
	"crawler/proto"
)

type XHS string

const (
	XHSName     = "小红书"
	UnknownName = "未知名称"
)

// ConvertSite 转换proto中的site为domain的site
func ConvertSite(site proto.Site) domain.Site {
	switch site {
	case proto.Site_XIAOHONGSHU:
		return XHSName
	default:
		return UnknownName
	}
}

func (x *XHS) GetSiteName(site proto.Site) string {
	switch site {
	case proto.Site_XIAOHONGSHU:
		return XHSName
	default:
		return UnknownName
	}
}
