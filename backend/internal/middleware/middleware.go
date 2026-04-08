package middleware

import (
	"strings"

	"xykitchen/backend/internal/config"
	"xykitchen/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CtxUser struct {
	ID       int
	Username string
	Role     string
	Realm    string // "" main app, "outlet" for outlet
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token := strings.TrimPrefix(h, "Bearer ")
		if token == "" {
			c.JSON(401, gin.H{"code": 401, "message": "未登录"})
			c.Abort()
			return
		}
		claims, err := services.ParseJWT(cfg, token)
		if err != nil {
			c.JSON(401, gin.H{"code": 401, "message": "Token已过期"})
			c.Abort()
			return
		}
		c.Set("user", CtxUser{ID: claims.ID, Username: claims.Username, Role: claims.Role, Realm: claims.Realm})
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := c.Get("user")
		if !ok {
			c.JSON(401, gin.H{"code": 401, "message": "未登录"})
			c.Abort()
			return
		}
		if cu, ok := u.(CtxUser); ok && cu.Role == "admin" {
			c.Next()
			return
		}
		c.JSON(403, gin.H{"code": 403, "message": "无管理员权限"})
		c.Abort()
	}
}

func OutletAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token := strings.TrimPrefix(h, "Bearer ")
		if token == "" {
			c.JSON(401, gin.H{"code": 401, "message": "未登录"})
			c.Abort()
			return
		}
		claims, err := services.ParseJWT(cfg, token)
		if err != nil {
			c.JSON(401, gin.H{"code": 401, "message": "Token已过期"})
			c.Abort()
			return
		}
		if claims.Realm != "outlet" {
			c.JSON(403, gin.H{"code": 403, "message": "非服务商账号"})
			c.Abort()
			return
		}
		c.Set("user", CtxUser{ID: claims.ID, Username: claims.Username, Role: claims.Role, Realm: claims.Realm})
		c.Next()
	}
}

func GetUser(c *gin.Context) (CtxUser, bool) {
	v, ok := c.Get("user")
	if !ok {
		return CtxUser{}, false
	}
	u, ok := v.(CtxUser)
	return u, ok
}
