package server

import (
	"context"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/util/auth"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	CookieSessionName = `SESSION`
)

var log = logger.Log.With(zap.String("cat", "auth"))

func JWTAuth(c *gin.Context) {
	cookie, err := c.Cookie(CookieSessionName)
	if err != nil {
		log.Debug("session cookie not found in request")
		login(c)
		return
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(cookie, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(auth.SecretKey), nil
	})
	if err != nil {
		log.Debug("parsing jwt with claims", zap.Error(err))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if !token.Valid {
		log.Debug("token is not valid")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	log.Debug(fmt.Sprintf("user id from token = %s", claims.Subject))

	ctx := c.Request.Context()
	newCtx := context.WithValue(ctx, "userID", claims.Subject)
	req := c.Request.WithContext(newCtx)
	c.Request = req

	log.Debug(fmt.Sprintf("request with user sub = %s", claims.Subject))
	c.Next()
}

func login(c *gin.Context) {
	tokenString, err := auth.GenerateToken()
	if err != nil {
		log.Error("build token", zap.Error(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.SetCookie(CookieSessionName, tokenString, int(auth.TokenExp/time.Second), "/",
		config.Get().Host(), true, true)
	c.AbortWithStatus(http.StatusUnauthorized)
}
