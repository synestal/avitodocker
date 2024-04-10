package post

import (
	"database/sql"
	"github.com/lib/pq"
)

type BannerId struct {
	ID string `json:"banner_id"`
}

func CreateNemBannerStorage(db *sql.DB, active string, content []string) (*BannerId, error) {
	query := `
    INSERT INTO
        banners_storage(title_banner, text_banner, url_banner, banner_state)
    VALUES ($1, $2, $3, $4)
    RETURNING id_banner
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(content[0], content[1], content[2], active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bannerId BannerId
	if rows.Next() {
		err := rows.Scan(&bannerId.ID)
		if err != nil {
			return nil, err
		}
	}

	return &bannerId, nil
}

func UpdateBannersStorage(db *sql.DB, id, active string, content []string) (string, error) {

	query := `
		UPDATE banners_storage SET banner_state = $2, title_banner = $3, text_banner = $4, url_banner = $5 WHERE id_banner = $1
		RETURNING CASE WHEN EXISTS (SELECT 1 FROM banners_storage WHERE id_banner = $1) THEN 'updated' ELSE NULL END AS result;
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	rows, err := stmt.Query(id, active, content[0], content[1], content[2])
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var ans string
	if rows.Next() {
		err := rows.Scan(&ans)
		if err != nil {
			return "", err
		}
	}

	return ans, nil
}

func DeleterBanners(db *sql.DB, item string) (string, error) {
	query := `WITH deleted AS (
    DELETE FROM banners_storage 
    WHERE id_banner = $1 
    RETURNING id_banner
)
SELECT 
    CASE 
        WHEN EXISTS (SELECT * FROM deleted) THEN 'deleted'
        ELSE 'NULL' 
    END AS result;
    	`

	stmt, err := db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	rows, err := stmt.Query(item)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var ans string
	if rows.Next() {
		err := rows.Scan(&ans)
		if err != nil {
			return "", err
		}
	}

	return ans, nil
}

func DeleterBannersPostponed(db *sql.DB, item []int) error {
	query := `
    INSERT INTO delayed_deletions (table_name, id_item)
    SELECT 'banners_storage', unnest($1::int[]);
    `

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(pq.Array(item))
	if err != nil {
		return err
	}

	return nil

}
