package server

import (
	"context"
	"fmt"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	pb "github.com/denis-oreshkevich/shortener/internal/app/server/proto"
	"github.com/denis-oreshkevich/shortener/internal/app/shortener"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedShortenerServer
	sh         *shortener.Shortener
	conf       config.Conf
	delChannel chan model.BatchDeleteEntry
}

func NewGRPCServer(sh *shortener.Shortener, conf config.Conf,
	delChannel chan model.BatchDeleteEntry) *GRPCServer {
	return &GRPCServer{
		sh:         sh,
		conf:       conf,
		delChannel: delChannel,
	}
}

func (gs *GRPCServer) CreateShortURL(ctx context.Context,
	req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	ctx = context.WithValue(ctx, model.UserIDKey{}, req.GetUserId())
	id, err := gs.sh.SaveURL(ctx, req.GetUrl())
	if err != nil {
		logger.Log.Error("sh.SaveURL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	url := fmt.Sprintf("%s/%s", gs.conf.BaseURL(), id)
	return &pb.CreateShortURLResponse{Result: url}, nil
}

func (gs *GRPCServer) BatchCreateShortURL(ctx context.Context,
	req *pb.BatchCreateShortURLRequest) (*pb.BatchCreateShortURLResponse, error) {
	var items []model.BatchReqEntry
	for _, item := range req.GetRecords() {
		items = append(items, model.BatchReqEntry{OriginalURL: item.OriginalUrl, CorrelationID: item.CorrelationId})
	}
	ctx = context.WithValue(ctx, model.UserIDKey{}, req.GetUserId())
	res, err := gs.sh.SaveURLBatch(ctx, items)
	if err != nil {
		logger.Log.Error("sh.SaveURLBatch", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var result []*pb.BatchCreateShortURLResponseData
	for _, item := range res {
		result = append(
			result,
			&pb.BatchCreateShortURLResponseData{
				ShortUrl:      item.ShortURL,
				CorrelationId: item.CorrelationID,
			})
	}
	return &pb.BatchCreateShortURLResponse{Records: result}, nil
}

func (gs *GRPCServer) GetByShort(ctx context.Context,
	req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	originalURL, err := gs.sh.FindURL(ctx, req.GetUrl())
	if err != nil {
		logger.Log.Error("sh.FindURL", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.GetOriginalURLResponse{OriginalUrl: originalURL}, nil
}

func (gs *GRPCServer) GetUserURLs(ctx context.Context,
	req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	ctx = context.WithValue(ctx, model.UserIDKey{}, req.GetUserId())
	urls, err := gs.sh.FindUserURLs(ctx)
	if err != nil {
		logger.Log.Error("sh.FindUserURLs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var result []*pb.ShortenData
	for _, item := range urls {
		result = append(result, &pb.ShortenData{ShortUrl: item.ShortURL, OriginalUrl: item.OriginalURL})
	}
	return &pb.GetUserURLsResponse{Records: result}, nil
}

func (gs *GRPCServer) DeleteUserURLsBatch(ctx context.Context,
	req *pb.DeleteUserURLsBatchRequest) (*pb.DeleteUserURLsBatchResponse, error) {
	ctx = context.WithValue(ctx, model.UserIDKey{}, req.GetUserId())
	entry := model.NewBatchDeleteEntry(req.GetUserId(), req.Urls)
	gs.delChannel <- entry
	gs.sh.DeleteUserURLs(ctx, gs.delChannel)
	return &pb.DeleteUserURLsBatchResponse{}, nil
}

func (gs *GRPCServer) GetStats(ctx context.Context,
	req *pb.ServiceStatsRequest) (*pb.ServiceStatsResponse, error) {
	stats, err := gs.sh.FindStats(ctx)
	if err != nil {
		logger.Log.Error("sh.FindStats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.ServiceStatsResponse{Urls: int64(stats.URLs), Users: int64(stats.Users)}, nil
}
