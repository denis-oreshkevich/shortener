package server

import (
	"context"
	"fmt"
	"net/http/httptest"

	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/gin-gonic/gin"
)

func ExampleServer_Get() {
	//Initializing storage
	s := storage.NewMapStorage()
	ctx := context.Background()
	userID := generator.UUIDString()
	ctx = context.WithValue(ctx, model.UserIDKey{}, userID)

	shortURL, err := s.SaveURL(ctx, userID, "http://localhost:30000")
	if err != nil {
		fmt.Println(fmt.Errorf("SaveURL : %w", err))
		return
	}
	//Initializing channel to delete URLs
	delChannel := make(chan model.BatchDeleteEntry, 3)
	//Initializing shortener
	sh := shortener.New(s)
	//Initializing config
	conf := config.Get()

	//Initializing sever entity
	srv := New(conf, sh, delChannel)

	//Initializing Gin server
	r := gin.New()

	// Initializing route to get original URL
	r.GET("/:id", srv.Get)

	// Creating new HTTP-request
	req := httptest.NewRequest("GET", conf.BaseURL()+"/"+shortURL, nil)
	w := httptest.NewRecorder()

	// Making request
	r.ServeHTTP(w, req)

	//Prints response body
	//fmt.Println(w.Body.String())

	//Prints response status
	fmt.Println(w.Code)

	// Output:
	// 307
}
