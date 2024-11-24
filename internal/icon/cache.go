package icon

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"sync"
)

type Cache struct {
	icons map[string]fyne.Resource
	mutex sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		icons: make(map[string]fyne.Resource),
	}
}

func (c *Cache) Get(path string) (fyne.Resource, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res, exists := c.icons[path]
	return res, exists
}

func (c *Cache) Set(path string, resource fyne.Resource) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.icons[path] = resource
}

func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.icons = make(map[string]fyne.Resource)
}

// 新增：获取图标的方法
func (c *Cache) GetIconResource(absPath string) fyne.Resource {
	// 先检查缓存
	if resource, exists := c.Get(absPath); exists {
		return resource
	}

	// 缓存中不存在，创建新的图标
	img, err := GetFileIcon2Image(absPath)
	if err != nil || img == nil {
		return theme.FileIcon()
	}

	resource, err := ImageToResource(img)
	if err != nil || resource == nil {
		return theme.FileIcon()
	}

	// 将新创建的图标存入缓存
	c.Set(absPath, resource)

	return resource
}
