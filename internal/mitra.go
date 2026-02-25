package mitra

import "mitra/internal/config"

type Mitra struct {
	Cfg *config.Config
}

func New() (*Mitra, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Mitra{Cfg: cfg}, nil
}
