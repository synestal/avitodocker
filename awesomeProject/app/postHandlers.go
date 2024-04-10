package app

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

func isNumeric(s string) bool {
	matched, _ := regexp.MatchString("^[0-9]+$", s)
	return matched
}

func allNumeric(slice []string) bool {
	for _, str := range slice {
		if !isNumeric(str) {
			return false
		}
	}
	return true
}

func PostAdminStateHandler(db *sql.DB) gin.HandlerFunc { // POST: set_admin
	return func(c *gin.Context) {
		id := c.Query("id")
		state := c.Query("state")
		err := SetadminState(db, id, state)
		var errJSONResponse ErrJSONResponse
		if err != nil {
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
			c.JSON(200, errJSONResponse)
		}
		errJSONResponse.ErrorJSON = "200, ОК"
		c.JSON(200, errJSONResponse)
	}
}

func CreateNewBannerHandler(db *sql.DB) gin.HandlerFunc {
	// http://localhost:8080/banner?admin_token=10&feature_id=15&tag_ids=22,12&content=notebooklovers,simpledescr,http://aboba.com&is_active=true
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		feature := c.Query("feature_id")
		tags := c.Query("tag_ids")
		content := c.Query("content")
		active := c.Query("is_active")
		parsedContent := strings.Split(content, ",")
		parsedTags := strings.Split(tags, ",")

		if len(parsedTags) < 2 || isNumeric(token) == false || isNumeric(feature) == false || allNumeric(parsedTags) == false || content == "" || active != "true" && active != "false" {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}

		request, BannerId, err := createNewBanner(db, token, feature, active, parsedContent, parsedTags)
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
		c.JSON(request, BannerId)
	}
}

func ChangeBannerHandler(db *sql.DB) gin.HandlerFunc {
	// http://localhost:8080/banner/%7Bid%7D?admin_token=10&feature_id=100&tag_ids=100,101&content=avitolovers,descr,http://avito.com&is_active=true&id=3
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		bannerid := c.Query("id")
		feature := c.Query("feature_id")
		tags := c.Query("tag_ids")
		content := c.Query("content")
		active := c.Query("is_active")
		parsedContent := strings.Split(content, ",")
		parsedTags := strings.Split(tags, ",")

		if len(parsedTags) < 2 || isNumeric(bannerid) == false || isNumeric(token) == false || isNumeric(feature) == false || allNumeric(parsedTags) == false || content == "" || active != "true" && active != "false" {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}
		request, err := changeBanner(db, token, bannerid, feature, active, parsedContent, parsedTags)
		var errJSONResponse ErrJSONResponse
		if err != nil {

		}
		if request == 500 {
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
		} else if request == 403 {
			errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
		} else if request == 401 {
			errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
		} else if request == 404 {
			errJSONResponse.ErrorJSON = "404, баннер для тега не найден"
		} else {
			errJSONResponse.ErrorJSON = "200, ОК"
		}
		c.JSON(request, errJSONResponse)
	}
}

func DeleteBannerHandler(db *sql.DB) gin.HandlerFunc {
	// http://localhost:8080/delete?id=3&admin_token=10
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		bannerid := c.Query("id")
		if isNumeric(bannerid) == false || isNumeric(token) == false {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}
		request, err := deleteBanner(db, token, bannerid)
		var errJSONResponse ErrJSONResponse
		if err != nil {

		}
		if request == 500 {
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
		} else if request == 403 {
			errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
		} else if request == 401 {
			errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
		} else if request == 404 {
			errJSONResponse.ErrorJSON = "404, баннер для тега не найден"
		} else {
			errJSONResponse.ErrorJSON = "204, Баннер успешно удален"
		}
		c.JSON(request, errJSONResponse)
	}
}

func DeleteFeatureTagHandler(db *sql.DB) gin.HandlerFunc {
	// http://localhost:8080/delete?id=3&admin_token=10
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		feature := c.Query("feature_id")
		tag := c.Query("tag_id")
		limit := c.Query("content")
		offset := c.Query("offset")
		if isNumeric(token) == false || feature == "" && tag == "" {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}

		request, err := deleteBannerByFeatureOrTag(db, token, feature, limit, offset, tag)
		var errJSONResponse ErrJSONResponse
		if err != nil {

		}
		if request == 500 {
			errJSONResponse.ErrorJSON = "500, внутренняя ошибка сервера"
		} else if request == 403 {
			errJSONResponse.ErrorJSON = "403, пользователь не имеет доступа"
		} else if request == 401 {
			errJSONResponse.ErrorJSON = "401, пользователь не авторизован"
		} else if request == 404 {
			errJSONResponse.ErrorJSON = "404, баннер для тега не найден"
		} else {
			errJSONResponse.ErrorJSON = "200, Баннер успешно удален"
		}
		c.JSON(request, errJSONResponse)
	}
}

func ChangeBannersHistoryHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("admin_token")
		id := c.Query("id")
		number := c.Query("number")
		if isNumeric(token) == false || isNumeric(number) == false || isNumeric(id) == false {
			var errJSONResponse ErrJSONResponse
			errJSONResponse.ErrorJSON = "400, некорректные данные"
			c.JSON(400, errJSONResponse)
			return
		}
		request, err := changeBannersHistory(db, token, number, id)
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
		errJSONResponse.ErrorJSON = "200, ОК"
		c.JSON(request, errJSONResponse)
	}
}
