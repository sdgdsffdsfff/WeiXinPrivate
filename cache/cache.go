package cache

import (
	"log"
	"time"

	"github.com/pmylund/go-cache"
)

type WCache struct {
	server   *cache.Cache
	filePath string
}

/**
 * 初始化Cache 加载cache文件
 */
func (w *WCache) InitCache(path string, defaultExpiration, cleanupInterval time.Duration) {
	w.server = cache.New(defaultExpiration, cleanupInterval)
	w.filePath = path
	// 从文件加载缓存key-value
	w.server.LoadFile(path)
}

/**
 * 启动监控 每10秒保存一次cache到文件
 */
func (w *WCache) RunSaveMonitor(sleepTime time.Duration) {
	go func() {
		for {
			time.Sleep(sleepTime)
			w.server.SaveFile(w.filePath)
			log.Println("save monitor: save to file done...")
		}
	}()
}

/**
 * 添加到cache 如果key不存在
 */
func (w *WCache) Add(k string, x interface{}, d time.Duration) error {
	return w.server.Add(k, x, d)
}

/**
 * 设置cache 如果key存在且不是永不过期
 */
func (w *WCache) Set(k string, x interface{}, d time.Duration) {
	w.server.Set(k, x, d)
}

/**
 * 替换cache
 */
func (w *WCache) Replace(k string, x interface{}, d time.Duration) error {
	return w.server.Replace(k, x, d)
}

/**
 * 获取cache
 */
func (w *WCache) Get(k string) (interface{}, bool) {
	return w.server.Get(k)
}

/**
 * 删除cache
 */
func (w *WCache) Delete(k string) {
	w.server.Delete(k)
}

/**
 * 删除过期cache
 */
func (w *WCache) DeleteExpired() {
	w.server.DeleteExpired()
}
