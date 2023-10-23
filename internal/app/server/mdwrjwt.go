package server

import (
	"context"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/util/auth"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
)

const (
	CookieName = `jwt-cookie`
)

func JWTAuth(c *gin.Context) {
	log := logger.Log.With(zap.String("cat", "auth"))
	ctx := c.Request.Context()
	tokenString, err := c.Cookie(CookieName)
	if err != nil {
		log.Debug("session cookie not found in request")
		tokenString, err = login(c)
		if err != nil {
			log.Error("login", zap.Error(err))
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx = context.WithValue(ctx, model.IsUserNew{}, true)
	}
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
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

	newCtx := context.WithValue(ctx, model.UserIDKey{}, claims.Subject)
	req := c.Request.WithContext(newCtx)
	c.Request = req

	log.Debug(fmt.Sprintf("request with user sub = %s", claims.Subject))
	c.Next()
}

func login(c *gin.Context) (string, error) {
	tokenString, err := auth.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token. %w", err)
	}
	c.SetCookie(CookieName, tokenString, int(auth.TokenExp.Seconds()), "",
		"", false, true)
	return tokenString, nil
}
