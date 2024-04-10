package app

import (
	"database/sql"
	"strings"
)

func SetadminState(db *sql.DB, id, state string) error {
	query := `
    INSERT INTO
        user_tokens(id, token_state)
    VALUES ($1, $2)
    ON CONFLICT (id)
	DO UPDATE SET token_state = EXCLUDED.token_state;
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	_, err = stmt.Query(id, state)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	return nil
}

func createNewBanner(db *sql.DB, token, feature, active string, content, tags []string) (int, *BannerId, error) {
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 500, nil, err
	}
	if avaliable == false {
		return 401, nil, err
	}
	if adminState == false {
		return 403, nil, err
	}
	bannerId, err := CreateNemBannerStorage(db, active, content)
	if err != nil {
		return 500, nil, err
	}
	err = CreateNewFeatureStorage(db, feature)
	if err != nil {
		return 500, nil, err
	}
	err = CreateNewTagStorage(db, feature, bannerId.ID, tags)
	if err != nil {
		return 500, nil, err
	}

	return 201, bannerId, nil
}

func changeBanner(db *sql.DB, token, bannerid, feature, active string, content, tags []string) (int, error) {
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}

	ans, err := UpdateBannersStorage(db, bannerid, active, content)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	ans, err = UpdateFeatureTagStorage(db, bannerid, feature, tags)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	return 200, nil
}

func deleteBanner(db *sql.DB, token, bannerid string) (int, error) {
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}

	ans, err := DeleterTags(db, bannerid)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	ans, err = DeleterBanners(db, bannerid)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	return 204, nil
}

func deleteBannerByFeatureOrTag(db *sql.DB, token, feature, limit, offset, tag string) (int, error) {
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}
	ids := make([]int, 0, 1)
	if tag == "" {
		ids, err = GetBannerIdByFeature(db, feature, limit, offset)
		if err != nil {
			return 400, err
		}
	} else {
		ids, err = GetBannerIdByTag(db, tag, limit, offset)
		if err != nil {
			return 400, err
		}
	}

	tagsErrChan := make(chan error)
	bannersErrChan := make(chan error)

	go func() {
		tagsErrChan <- DeleterTagsPostponed(db, ids) // Вызываем функцию и передаем ошибку через канал
	}()

	go func() {
		bannersErrChan <- DeleterBannersPostponed(db, ids) // Вызываем функцию и передаем ошибку через канал
	}()

	tagsErr := <-tagsErrChan
	bannersErr := <-bannersErrChan

	if tagsErr != nil {
		return 500, tagsErr
	}

	if bannersErr != nil {
		return 500, bannersErr
	}

	return 200, nil
}

func changeBannersHistory(db *sql.DB, token, number, id string) (int, error) {
	avaliable, adminState, err := GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}

	ans, err := ChangeHistoryBannersStorage(db, number, id)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	ans, err = ChangeHistoryFeatureTagStorage(db, number, id)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	return 200, nil
}

type TagsTable struct {
	Id       string `json:"id"`
	Features string `json:"features"`
	Tags     string `json:"tags"`
}

func ChangeHistoryFeatureTagStorage(db *sql.DB, number, id string) (string, error) {
	query := `SELECT id_banner, features_id, tag_list
FROM (
    SELECT 
        id_banner, 
        features_id, 
        tag_list,
        ROW_NUMBER() OVER (PARTITION BY id_banner ORDER BY id) AS row_num
    FROM 
        history_features
    WHERE 
        id_banner = $2
) AS subquery
WHERE 
    row_num = $1;
    	`
	stmt, err := db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	rows, err := stmt.Query(number, id)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	query = `DELETE FROM history_features
WHERE id_banner = $2
AND id = (
    SELECT id
    FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY id_banner ORDER BY id) AS row_num
        FROM history_features
        WHERE id_banner = $2
    ) AS subquery
    WHERE row_num = $1
);`

	stmt, err = db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	_, err = stmt.Query(number, id)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	var tagsTable TagsTable
	if rows.Next() {
		err := rows.Scan(&tagsTable.Id, &tagsTable.Features, &tagsTable.Tags)
		if err != nil {
			return "", err
		}
	}
	tagsTable.Tags = tagsTable.Tags[1 : len(tagsTable.Tags)-1]
	parsedTags := strings.Split(tagsTable.Tags, ",")
	ans, err := UpdateFeatureTagStorage(db, tagsTable.Id, tagsTable.Features, parsedTags)
	if err != nil {
		return "", err
	}
	return ans, nil
}

type Bannerstable struct {
	Id        string `json:"banner_id"`
	Banner    Banner `json:"content"`
	Flag      string `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ChangeHistoryBannersStorage(db *sql.DB, number, id string) (string, error) {
	query := `SELECT id_banner, title_banner, text_banner, url_banner, banner_state, created_at, updated_at
FROM (
    SELECT 
        id_banner, 
        title_banner, 
        text_banner, 
        url_banner, 
        banner_state, 
        created_at, 
        updated_at,
        ROW_NUMBER() OVER (PARTITION BY id_banner ORDER BY id) AS row_num
    FROM 
        history_banenrs
    WHERE 
        id_banner = $2
) AS subquery
WHERE 
    row_num = $1;
    	`
	stmt, err := db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	rows, err := stmt.Query(number, id)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	query = `DELETE FROM history_banenrs
WHERE id_banner = $2
AND id = (
    SELECT id
    FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY id_banner ORDER BY id) AS row_num
        FROM history_banenrs
        WHERE id_banner = $2
    ) AS subquery
    WHERE row_num = $1
);`

	stmt, err = db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	_, err = stmt.Query(number, id)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	var bannerstable Bannerstable
	if rows.Next() {
		err := rows.Scan(&bannerstable.Id, &bannerstable.Banner.Title, &bannerstable.Banner.Text, &bannerstable.Banner.Url, &bannerstable.Flag, &bannerstable.CreatedAt, &bannerstable.UpdatedAt)
		if err != nil {
			return "", err
		}
	}
	content := make([]string, 3, 3)
	content[0] = bannerstable.Banner.Title
	content[1] = bannerstable.Banner.Text
	content[2] = bannerstable.Banner.Url
	ans, err := UpdateBannersStorage(db, bannerstable.Id, bannerstable.Flag, content)
	if err != nil {
		return "", err
	}
	return ans, nil
}
