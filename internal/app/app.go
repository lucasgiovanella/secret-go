package app

import (
	"github.com/lucasgiovanella/secret-go/internal/auth"
	"github.com/lucasgiovanella/secret-go/internal/config"
	"github.com/lucasgiovanella/secret-go/internal/db"
	"github.com/lucasgiovanella/secret-go/internal/model"
)

type App struct {
	Config *config.Config
	Auth   *auth.Auth
	DB     *db.DB
	Store  *model.Store
}

func NewApp() (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	authManager, err := auth.NewAuth(cfg)
	if err != nil {
		return nil, err
	}

	dbManager, err := db.NewDB(cfg.StoreDBPath)
	if err != nil {
		return nil, err
	}

	store, err := dbManager.LoadStore()
	if err != nil || store == nil {
		store = model.NewStore()
	}

	return &App{
		Config: cfg,
		Auth:   authManager,
		DB:     dbManager,
		Store:  store,
	}, nil
}

func (a *App) SaveStore() error {
	return a.DB.SaveStore(a.Store)
}

func (a *App) Close() error {
	return a.DB.Close()
}
