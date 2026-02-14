package main

import (
	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/usecase"
	"auth-service/pkg/config"
	"database/sql"
	"fmt"

	"auth-service/pkg/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Server struct {
	// server fields
	authUseCase usecase.AuthUseCaseInterface
	tenantUseCase usecase.TenantUseCaseInterface
	engine      *gin.Engine
	host        string
	jwtService  service.JwtServiceImpl
}

func (s *Server) initRoute() {
	// api group
	apiGroup := s.engine.Group("")
	midware := middleware.NewAuthMiddleware(s.jwtService)
	handler.NewAuthHandler(s.authUseCase, s.jwtService, apiGroup, midware).Routes()
	handler.NewTenantHandler(s.tenantUseCase, s.jwtService, apiGroup, midware).Routes()

}

func (s *Server) Run() {
	s.initRoute()
	if err := s.engine.Run(s.host); err != nil {
		panic(fmt.Errorf("server not running on host %s, because error %v", s.host, err.Error()))
	}
}

func NewServer() *Server {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Println(err)
		panic("Config error")
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)
	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		panic("connection error")
	}

	// service
	jwtService := service.NewJwtService(cfg.TokenConfig)

	//init repository
	repo_user := repository.NewUserRepository(db)
	repo_audit_log := repository.NewAuditLogRepository(db)
	// := repository.NewPermissionRepository(db)
	repo_role := repository.NewRoleRepository(db)
	repo_tenant := repository.NewTenantRepository(db)
	repo_user_role := repository.NewUserRoleRepository(db)
	repo_refresh_token := repository.NewRefreshTokenRepository(db)

	// init usecase
	authUseCase := usecase.NewAuthUseCase(
		repo_user,
		repo_role,
		repo_user_role,
		repo_refresh_token,
		repo_audit_log,
		jwtService,
		cfg.RefreshTokenExpiry,
	)

	tenantUseCase := usecase.NewTenantUseCase(
		repo_tenant,
		repo_audit_log,
	)

	// HTTP Init
	host := fmt.Sprintf(":%s", cfg.ApiPort)
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	return &Server{
		authUseCase: authUseCase,
		tenantUseCase: tenantUseCase,
		engine:      engine,
		host:        host,
		jwtService:  jwtService,
	}
}
