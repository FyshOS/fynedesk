package icon

import "fyshos.com/fynedesk"

type appCache struct {
	source  fynedesk.ApplicationProvider
	appList []fynedesk.AppData
}

func (c *appCache) forEachCachedApplication(f func(string, fynedesk.AppData) bool) {
	if c.appList == nil {
		c.appList = c.source.AvailableApps()
	}

	for _, a := range c.appList {
		if f(a.Name(), a) {
			return
		}
	}
}

func newAppCache(c fynedesk.ApplicationProvider) *appCache {
	return &appCache{source: c}
}
