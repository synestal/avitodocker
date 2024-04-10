package http_server

import (
	"awesomeProject/config"
	getHandle "awesomeProject/internal/app/handlers/get"
	postHandle "awesomeProject/internal/app/handlers/post"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"log"
)

type Server struct {
	Cfg         *config.Config
	Router      *gin.Engine
	Db          *sql.DB
	RedisClient *redis.Client
}

func NewServer(cfg *config.Config) *Server {
	db, rd, err := InitDb(cfg)
	if err != nil {
		log.Fatalf("Cannot map handlers. Error: {%s}", err)
	}
	return &Server{
		Cfg:         cfg,
		Router:      gin.Default(),
		Db:          db,
		RedisClient: rd,
	}
}

func (s *Server) Run() error {
	s.Router.GET("/banners", getHandle.GetBannerHandler(s.Db, s.RedisClient))
	s.Router.GET("/set_admin", postHandle.PostAdminStateHandler(s.Db))
	s.Router.GET("/banner", postHandle.CreateNewBannerHandler(s.Db))
	s.Router.GET("/banner/{id}", postHandle.ChangeBannerHandler(s.Db))
	s.Router.GET("/delete", postHandle.DeleteBannerHandler(s.Db))
	s.Router.GET("/bannerget", getHandle.GetBannerByFilterHandler(s.Db))
	s.Router.GET("/featuretagdelete", postHandle.DeleteFeatureTagHandler(s.Db))
	s.Router.GET("/gethistory", getHandle.GetBannersHistoryHandler(s.Db))
	s.Router.GET("/changehistory", postHandle.ChangeBannersHistoryHandler(s.Db))

	if err := s.Router.Run(s.Cfg.Server.Port); err != nil {
		log.Fatalf("Cannot listen. Error: {%s}", err)
	}

	return nil
}
