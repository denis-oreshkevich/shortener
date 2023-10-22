package shortener

import (
	"context"
	"github.com/denis-oreshkevich/shortener/internal/app/config"
	"github.com/denis-oreshkevich/shortener/internal/app/model"
	"github.com/denis-oreshkevich/shortener/internal/app/storage"
)

type Shortener struct {
	conf    config.Conf
	storage storage.Storage
}

func New(conf config.Conf, st storage.Storage) *Shortener {
	return &Shortener{
		conf:    conf,
		storage: st,
	}
}

func (sh *Shortener) SaveURLBatch(ctx context.Context,
	batch []model.BatchReqEntry) ([]model.BatchRespEntry, error) {
	//TODO add business layer
	return nil, nil
}
