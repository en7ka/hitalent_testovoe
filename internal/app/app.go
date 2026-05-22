package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/en7ka/hitalent_testovoe/internal/config"
	"github.com/en7ka/hitalent_testovoe/internal/db"
	"github.com/en7ka/hitalent_testovoe/internal/repository"
	"github.com/en7ka/hitalent_testovoe/internal/service"
	httptransport "github.com/en7ka/hitalent_testovoe/internal/transport/http"
	"gorm.io/gorm"
)

type App struct {
	server *http.Server
	db     *gorm.DB
}

func New(cfg config.Config) (*App, error) {
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	departmentRepository := repository.NewDepartmentRepository(database)
	employeeRepository := repository.NewEmployeeRepository(database)
	transactor := repository.NewTransactor(database)

	departmentService := service.NewDepartmentService(departmentRepository, employeeRepository, transactor)
	employeeService := service.NewEmployeeService(employeeRepository, departmentRepository)

	router := httptransport.NewRouter(departmentService, employeeService)

	return &App{
		db: database,
		server: &http.Server{
			Addr:              ":" + cfg.HTTPPort,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (a *App) Run() error {
	log.Printf("server listening on %s", a.server.Addr)

	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	sqlDB, err := a.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
