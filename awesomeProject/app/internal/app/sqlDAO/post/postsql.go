package post

import (
	getsql "awesomeProject/internal/app/sqlDAO/get"
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
	defer stmt.Close()
	_, err = stmt.Query(id, state)
	if err != nil {
		return err
	}

	return nil
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
	defer stmt.Close()
	rows, err := stmt.Query(number, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

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
	_, err = stmt.Query(number, id)
	if err != nil {
		return "", err
	}

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
	Id        string        `json:"banner_id"`
	Banner    getsql.Banner `json:"content"`
	Flag      string        `json:"is_active"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
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
	defer stmt.Close()
	rows, err := stmt.Query(number, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()

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
	defer stmt.Close()
	_, err = stmt.Query(number, id)
	if err != nil {
		return "", err
	}

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
