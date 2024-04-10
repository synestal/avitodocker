package avito_test_trainee

import (
	"awesomeProject/config"
	server "awesomeProject/internal/http-server"
	_ "fmt"
	_ "github.com/lib/pq"
	"log"
	_ "time"
)

func main() {

	viperInstance, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Cannot load config. Error: {%s}", err.Error())
	}

	cfg, err := config.ParseConfig(viperInstance)
	if err != nil {
		log.Fatalf("Cannot parse config. Error: {%s}", err.Error())
	}

	s := server.NewServer(cfg)
	if err = s.Run(); err != nil {
		log.Print(err)
	}

}
