package get

import (
	get "awesomeProject/internal/app/acc/get-funcs"
	getsql "awesomeProject/internal/app/sqlDAO/get"
	cashersql "awesomeProject/internal/redis-casher"
	help "awesomeProject/pkg/func"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"net/http"
)

type ErrJSONResponse struct {
	ErrorJSON string `json:"error"`
}

type bannerIfc interface {
	GetBannerHandler(db *sql.DB, redisClient *redis.Client) gin.HandlerFunc
	GetBannerByFilterHandler(db *sql.DB) gin.HandlerFunc
	GetBannersHistoryHandler(db *sql.DB) gin.HandlerFunc
}

type BannerHandler struct {
	service bannerIfc
}

func NewBannerHandler(service bannerIfc) BannerHandler {
	return BannerHandler{
		service: service,
	}
}

func GetBannerHandler(db *sql.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// http://localhost:8080/banners?tag_id=11&feature_id=15&admin_token=10
		tagID := c.Query("tag_id")
		featureID := c.Query("feature_id")
		useLast := c.Query("use_last_revision")
		token := c.Query("admin_token")

		if help.IsNumeric(tagID) == false || help.IsNumeric(featureID) == false || useLast != "false" && useLast != "true" || help.IsNumeric(token) == false {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}

		useLastRevision := useLast == "true"
		var adminState bool
		var err error
		avaliable, adminState, err := getsql.GetAdminState(db, token)
		if err != nil {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
			c.JSON(500, errJSONResponse)
			return
		}

		// Проверка кэша
		if !useLastRevision {
			state, banner, err := cashersql.GetBannerFromCache(redisClient, db, tagID, featureID)
			if err != nil {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
				c.JSON(500, errJSONResponse)
				return
			}
			if avaliable == false {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
				c.JSON(401, errJSONResponse)
				return
			}
			if state == "false" && adminState == false {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
				c.JSON(403, errJSONResponse)
				return
			}
			if banner.Title == "" && banner.Text == "" && banner.Url == "" {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "404, баннер не найден"
				c.JSON(404, errJSONResponse)
				return
			}
			c.JSON(http.StatusOK, banner)
		}

		if useLastRevision {
			var state string
			banner, state, err := getsql.GetBannerFromDB(db, tagID, featureID)
			if err != nil {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
				c.JSON(500, errJSONResponse)
				return
			}
			if avaliable == false {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
				c.JSON(401, errJSONResponse)
				return
			}
			if state == "false" && adminState == false {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
				c.JSON(403, errJSONResponse)
				return
			}
			if banner.Title == "" && banner.Text == "" && banner.Url == "" {
				var errJSONResponse ErrJSONResponse
				errJSONResponse.ErrorJSON = "404, баннер не найден"
				c.JSON(404, errJSONResponse)
				return
			}
			c.JSON(http.StatusOK, banner)
		}
	}
}

func GetBannerByFilterHandler(db *sql.DB) gin.HandlerFunc {
	// http://localhost:8080/bannerget?admin_token=10&feature_id=15&content=5&offset=0
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		feature := c.Query("feature_id")
		tag := c.Query("tag_id")
		limit := c.Query("content")
		offset := c.Query("offset")

		if help.IsNumeric(token) == false || help.IsNumeric(feature) == false && help.IsNumeric(tag) == false {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}

		request, bannerList, err := get.GetBannerByFilter(db, token, feature, limit, offset, tag)
		if err != nil {

		}
		var errJSONResponse ErrJSONResponse
		if request == 500 {
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
			c.JSON(request, errJSONResponse)
			return
		} else if request == 403 {
			errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
			c.JSON(request, errJSONResponse)
			return
		} else if request == 401 {
			errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
			c.JSON(request, errJSONResponse)
			return
		}

		c.JSON(request, bannerList)
	}
}

func GetBannersHistoryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		id := c.Query("id")
		if help.IsNumeric(token) == false || help.IsNumeric(id) == false {
			var errJSONResponse ErrJSONResponse
			fmt.Println("ErrHere")
			c.JSON(400, errJSONResponse)
			return
		}
		request, bannerList, err := get.GetBannersHistory(db, token, id)
		if err != nil {

		}
		var errJSONResponse ErrJSONResponse
		if request == 500 {
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
			c.JSON(request, errJSONResponse)
			return
		} else if request == 403 {
			errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
			c.JSON(request, errJSONResponse)
			return
		} else if request == 401 {
			errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
			c.JSON(request, errJSONResponse)
			return
		}

		c.JSON(request, bannerList)
	}
}
