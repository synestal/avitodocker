package redis_casher

import (
	getsql "awesomeProject/internal/app/sqlDAO/get"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type BannerPlusState struct {
	Banner *getsql.Banner `json:"banner"`
	State  string         `json:"state"`
}

func GetBannerFromCache(redisClient *redis.Client, db *sql.DB, tagID, featureID string) (string, *getsql.Banner, error) {
	cacheKey := fmt.Sprintf("banner:%s:%s", tagID, featureID)

	cachedProducts, err := redisClient.Get(cacheKey).Bytes()
	bannerPlusState := BannerPlusState{}
	if err != nil {
		var state string
		dbProducts, state, err := getsql.GetBannerFromDB(db, tagID, featureID)
		if err != nil {
			return state, nil, err
		}
		bannerPlusState.Banner = dbProducts
		bannerPlusState.State = state
		cachedProducts, err = json.Marshal(bannerPlusState)
		if err != nil {
			return "", nil, err
		}
		err = redisClient.Set(cacheKey, cachedProducts, 300*time.Second).Err()
		if err != nil {
			return "", nil, err
		}
		fmt.Println("Cashing")
		return state, dbProducts, nil
	}

	fmt.Println("From cashe")
	err = json.Unmarshal([]byte(cachedProducts), &bannerPlusState)
	if err != nil {
		return "", nil, err
	}
	return bannerPlusState.State, bannerPlusState.Banner, nil
}
