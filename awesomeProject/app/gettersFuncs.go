package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/lib/pq"
	"sync"
	"time"
)

type Banner struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Url   string `json:"url"`
}

type FilteredBanner struct {
	Id         string `json:"banner_id"`
	TagIds     string `json:"tag_ids"`
	FeatureIds string `json:"feature_id"`
	Banner     Banner `json:"content"`
	Flag       string `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type HistoryBanner struct {
	Id         string `json:"change_id"`
	TagIds     string `json:"tag_ids"`
	FeatureIds string `json:"feature_id"`
	Banner     Banner `json:"content"`
	Flag       string `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type BannerState struct {
	state string
}

func intersection(slice1, slice2 []int) []int {
	set := make(map[int]bool)
	for _, num := range slice1 {
		set[num] = true
	}
	var result []int
	for _, num := range slice2 {
		if set[num] {
			result = append(result, num)
		}
	}

	return result
}

type BannerPlusState struct {
	Banner *Banner `json:"banner"`
	State  string  `json:"state"`
}

func GetBannerFromCache(redisClient *redis.Client, db *sql.DB, tagID, featureID string) (string, *Banner, error) {
	// Формирование ключа кэша
	cacheKey := fmt.Sprintf("banner:%s:%s", tagID, featureID)

	// Получение JSON-объекта из Redis
	cachedProducts, err := redisClient.Get(cacheKey).Bytes()
	bannerPlusState := BannerPlusState{}
	if err != nil {
		var state string
		dbProducts, state, err := GetBannerFromDB(db, tagID, featureID)
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

	// Десериализация JSON-объекта в Banner
	fmt.Println("From cashe")
	err = json.Unmarshal([]byte(cachedProducts), &bannerPlusState)
	if err != nil {
		return "", nil, err
	}
	return bannerPlusState.State, bannerPlusState.Banner, nil
}

func GetBannerFromDB(db *sql.DB, tagID, featureID string) (*Banner, string, error) {
	// http://localhost:8080/banners?tag_id=21&feature_id=19&tocken=10
	query := `
		WITH a AS (
    		SELECT tags.id_banner 
    		FROM tags 
    		WHERE features_id = $2 AND $1 = ANY(tag_list) 
    		LIMIT 1
		)
		SELECT title_banner, text_banner, url_banner, banner_state
			FROM banners_storage 
		JOIN a ON banners_storage.id_banner = a.id_banner;
    `
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	rows, err := stmt.Query(tagID, featureID)
	if err != nil {
		return nil, "", err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var banner Banner
	var bannerState BannerState
	if rows.Next() {
		err := rows.Scan(&banner.Title, &banner.Text, &banner.Url, &bannerState.state)
		if err != nil {
			return nil, "", err
		}
	}

	return &banner, bannerState.state, nil
}

type Admin struct {
	avaliable string
	admin     string
}

func GetAdminState(db *sql.DB, tocken string) (bool, bool, error) {
	query := `SELECT 
    CASE 
        WHEN EXISTS (SELECT 1 FROM user_tokens WHERE id = $1) THEN TRUE
        ELSE FALSE
    END AS found_subj,
    CASE 
        WHEN (SELECT token_state FROM user_tokens WHERE id = $1) THEN TRUE
        ELSE FALSE
    END AS status
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return false, false, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(tocken)
	if err != nil {
		return false, false, err
	}
	defer stmt.Close()

	var admin Admin
	if rows.Next() {
		err := rows.Scan(&admin.avaliable, &admin.admin)
		if err != nil {
			return false, false, err
		}
	}

	return admin.avaliable == "true", admin.admin == "true", nil
}

func GetBannerIdByTag(db *sql.DB, tag, limit, offset string) ([]int, error) {
	var ids []int

	query := `
    SELECT
        id_banner
    FROM tags
    WHERE $1 = ANY(tag_list)
`
	if limit != "" {
		query += "LIMIT " + limit + " "
	}
	if offset != "" {
		query += "OFFSET " + offset
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		return ids, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(tag)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func GetBannerIdByFeature(db *sql.DB, feature, limit, offset string) ([]int, error) {
	var ids []int

	query := `
    SELECT
        id_banner
    FROM tags
    WHERE $1 = features_id
`
	if limit != "" {
		query += "LIMIT " + limit + " "
	}
	if offset != "" {
		query += "OFFSET " + offset
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return ids, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(feature)
	if err != nil {
		return ids, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func GetBannerStorage(db *sql.DB, ids []int, filteredBanner *[]FilteredBanner) error {
	query := `
    SELECT
        bs.id_banner, bs.title_banner, bs.text_banner, bs.url_banner, bs.banner_state, bs.created_at, bs.updated_at, t.tag_list, t.features_id 
    FROM banners_storage bs 
    JOIN 
    	tags t 
	ON 
    bs.id_banner = t.id_banner
    WHERE bs.id_banner = ANY($1);
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var temp FilteredBanner
		err := rows.Scan(&temp.Id, &temp.Banner.Title, &temp.Banner.Text, &temp.Banner.Url, &temp.Flag, &temp.CreatedAt, &temp.UpdatedAt, &temp.TagIds, &temp.FeatureIds)
		if err != nil {
			return err
		}
		*filteredBanner = append(*filteredBanner, temp)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func GetBannerByFilter(db *sql.DB, token, feature, limit, offset, tag string) (int, []FilteredBanner, error) {
	filteredBanner := make([]FilteredBanner, 0, 1)
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 400, filteredBanner, err
	}
	if avaliable == false {
		return 401, filteredBanner, err
	}
	if adminState == false {

		return 403, filteredBanner, err // Пользователь не имеет доступа
	}
	if tag == "" {
		ids, err := GetBannerIdByFeature(db, feature, limit, offset)
		if err != nil {
			return 400, filteredBanner, err
		}
		err = GetBannerStorage(db, ids, &filteredBanner)
		fmt.Println(filteredBanner)
		if err != nil {
			return 400, filteredBanner, err
		}
	} else if feature == "" {
		ids, err := GetBannerIdByTag(db, tag, limit, offset)
		if err != nil {
			return 400, filteredBanner, err
		}
		err = GetBannerStorage(db, ids, &filteredBanner)
		if err != nil {
			return 400, filteredBanner, err
		}
	} else {
		var (
			idsFea []int
			idsTag []int
			err    error
		)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			idsFea, err = GetBannerIdByFeature(db, feature, limit, offset)
			if err != nil {

			}
		}()
		go func() {
			defer wg.Done()
			idsTag, err = GetBannerIdByTag(db, tag, limit, offset)
			if err != nil {

			}
		}()
		wg.Wait()

		ids := intersection(idsFea, idsTag)
		err = GetBannerStorage(db, ids, &filteredBanner)
		if err != nil {
			return 400, filteredBanner, err
		}
	}
	return 200, filteredBanner, nil
}

func GetBannerHistoryStorage(db *sql.DB, id string, filteredBanner *[]HistoryBanner) error {
	query := `
SELECT 
    title_banner, text_banner, url_banner, banner_state, created_at, updated_at
FROM history_banenrs WHERE id_banner = $1;
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return err
	}
	defer rows.Close()

	query = `
SELECT 
    id, features_id, tag_list
FROM history_features WHERE id_banner = $1;
    `

	stmtE, err := db.Prepare(query)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer stmtE.Close()

	rowsE, err := stmtE.Query(id)
	if err != nil {
		return err
	}
	defer rowsE.Close()

	for rows.Next() && rowsE.Next() {
		var temp HistoryBanner
		err := rows.Scan(&temp.Banner.Title, &temp.Banner.Text, &temp.Banner.Url, &temp.Flag, &temp.CreatedAt, &temp.UpdatedAt)
		if err != nil {
			return err
		}
		err = rowsE.Scan(&temp.Id, &temp.FeatureIds, &temp.TagIds)
		if err != nil {
			return err
		}
		*filteredBanner = append(*filteredBanner, temp)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func GetBannersHistory(db *sql.DB, token, id string) (int, []HistoryBanner, error) {
	filteredBanner := make([]HistoryBanner, 0, 1)
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 400, filteredBanner, err
	}
	if avaliable == false {
		return 401, filteredBanner, err
	}
	if adminState == false {
		return 403, filteredBanner, err // Пользователь не имеет доступа
	}
	err = GetBannerHistoryStorage(db, id, &filteredBanner)
	if err != nil {
		return 400, filteredBanner, err
	}
	return 200, filteredBanner, nil
}
