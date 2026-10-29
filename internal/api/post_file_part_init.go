package api

import (
	"fmt"
	//"io"
	//"net/url"
	//"os"
	//"path"
	//"regexp"
	//"strings"

	//"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	//"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
	"github.com/jellydator/ttlcache/v3"
)

var CacheUniqueKeyTTL *ttlcache.Cache[string, string]

func PostFilePartInit(c fiber.Ctx) error {

	rndKey := util.RandStringRunes(32)

	CacheUniqueKeyTTL.Set(rndKey, "1", ttlcache.DefaultTTL)

	fmt.Println("PostFilePartInit key=", rndKey)

	return c.JSON(fiber.Map{
		"code": 200,
		"key":  rndKey,
	})
}
