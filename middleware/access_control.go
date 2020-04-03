// middleware/access_control.go
package middleware

import (
	"errors"
	"fmt"
	"log"
	"web2/config"
	"web2/handler"

	"github.com/allegro/bigcache"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Authenticate determines if current subject has logged in.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session id
		sessionId, _ := c.Cookie("current_subject")
		// Get current subject
		sub, err := handler.GlobalCache.Get(sessionId)
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			c.AbortWithStatusJSON(401, config.RestResponse{Message: "user hasn't logged in yet"})
			return
		}
		c.Set("current_subject", string(sub))
		c.Next()
	}
}

// middleware/access_control.go

// Authorize determines if current subject has been authorized to take an action on an object.
func Authorize(obj string, act string, adapter *gormadapter.Adapter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current user/subject
		val, existed := c.Get("current_subject")
		if !existed {
			c.AbortWithStatusJSON(401, config.RestResponse{Message: "user hasn't logged in yet"})
			return
		}
		// Casbin enforces policy
		ok, err := enforce(val.(string), obj, act, adapter)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(500, config.RestResponse{Message: "error occurred when authorizing user"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(403, config.RestResponse{Message: "forbidden"})
			return
		}
		c.Next()
	}
}

func enforce(sub string, obj string, act string, adapter *gormadapter.Adapter) (bool, error) {
	// Load model configuration file and policy store adapter
	enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	if err != nil {
		return false, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}
	// Load policies from DB dynamically
	err = enforcer.LoadPolicy()
	if err != nil {
		return false, fmt.Errorf("failed to load policy from DB: %w", err)
	}
	// Verify
	ok, err := enforcer.Enforce(sub, obj, act)
	return ok, err
}
