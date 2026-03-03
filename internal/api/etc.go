package api

import (
	//"strconv"
	"sync"
	"time"

	//"github.com/western/http-here/v2/internal/model2"

	"github.com/gofiber/fiber/v2"
)

type FileRow struct {
	IsDir    bool
	FullPath string
	Name     string

	Size      int64
	SizeHuman string

	ModTime      time.Time
	ModTimeHuman string
	Md5          string

	IsPreviewImg bool
	IsPreviewDoc bool
	IsEditDoc    bool
	IsEditCode   bool
	IsEditMd     bool
	Rndm         string
}

func setCookie(c *fiber.Ctx, name, value string) {
	cookie := new(fiber.Cookie)
	cookie.Name = name
	cookie.Value = value
	c.Cookie(cookie)
}

// -------------------------------------------------------------------------------------------------------------------------------------------

/*
type ErrorDescription struct {
	C          *fiber.Ctx
	DB         *gorm.DB
	StatusCode int
	Tag        string
	Title      string
	MsgShort   string
	MsgLong    string
}
*/

/*
func handleErrorJSON(c *fiber.Ctx, db *gorm.DB, statusCode int, tag, msgShort, msgLong string) error {

	model2.EventLogAdd(c, strconv.Itoa(statusCode), tag, msgLong)

	return c.Status(statusCode).JSON(fiber.Map{
		"code": statusCode,
		"msg":  msgShort,
	}, "application/json")
}
*/

/*
func handleErrorJSON(ed ErrorDescription) error {

	model.EventLogAdd(ed.DB, ed.C, strconv.Itoa(ed.StatusCode), ed.Tag, ed.MsgLong)

	return ed.C.Status(ed.StatusCode).JSON(fiber.Map{
		"code": ed.StatusCode,
		"msg":  ed.MsgShort,
	}, "application/json")
}
*/

/*
func handleErrorHTML(c *fiber.Ctx, db *gorm.DB, statusCode int, tag, title, msgShort, msgLong string) error {

	model2.EventLogAdd(c, strconv.Itoa(statusCode), tag, msgLong)

	return c.Status(statusCode).Render("view/error", fiber.Map{
		"Title":   title,
		"Message": msgShort,
	}, "view/layout/error")
}
*/

/*
func handleErrorHTML(ed ErrorDescription) error {

	model.EventLogAdd(ed.DB, ed.C, strconv.Itoa(ed.StatusCode), ed.Tag, ed.MsgLong)

	return ed.C.Status(ed.StatusCode).Render("view/error", fiber.Map{
		"Title":   ed.Title,
		"Message": ed.MsgShort,
	}, "view/layout/error")
}
*/

// -------------------------------------------------------------------------------------------------------------------------------------------

var (
	dirCache    = make(map[string]*cachedDir)
	dirCacheMux sync.RWMutex
	//cacheTimeout = 30 * time.Second
)

type cachedDir struct {
	rows_dir  []FileRow
	rows_file []FileRow
	timestamp time.Time
}

func (cd *cachedDir) isExpired(sc int) bool {

	return time.Since(cd.timestamp) > time.Duration(sc)*time.Second
}

func CleanupCacheDir(sc int) {

	dirCacheMux.Lock()
	for key, cached := range dirCache {
		if cached.isExpired(sc) {

			delete(dirCache, key)
		}
	}
	dirCacheMux.Unlock()
}

func RemoveCacheDir(readTarget string) bool {
	dirCacheMux.RLock()
	_, exists := dirCache[readTarget]
	if exists {
		delete(dirCache, readTarget)
		dirCacheMux.RUnlock()
		return true
	}

	dirCacheMux.RUnlock()
	return false
}

// -------------------------------------------------------------------------------------------------------------------------------------------

func GetBoolFromLocals(c *fiber.Ctx, key string) bool {
	if val, ok := c.Locals(key).(bool); ok {
		return val
	}
	return false
}

func GetStringFromLocals(c *fiber.Ctx, key string, defaultValue string) string {
	if val, ok := c.Locals(key).(string); ok {
		return val
	}
	return defaultValue
}

func GetIntFromLocals(c *fiber.Ctx, key string, defaultValue int) int {
	if val, ok := c.Locals(key).(int); ok {
		return val
	}
	return defaultValue
}

// -------------------------------------------------------------------------------------------------------------------------------------------
