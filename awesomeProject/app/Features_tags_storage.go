package app

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

func CreateNewFeatureStorage(db *sql.DB, feature string) error {
	query := `INSERT INTO features (features)
VALUES ($1)
ON CONFLICT (features)
DO NOTHING;
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

	_, err = stmt.Query(feature)
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

func CreateNewTagStorage(db *sql.DB, feature, bannerId string, tags []string) error {
	query := `
    INSERT INTO tags (id_banner, tag_list, features_id) 
	VALUES ($1, $2, $3);
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

	_, err = stmt.Query(bannerId, pq.Array(tags), feature)
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

func UpdateFeatureTagStorage(db *sql.DB, id, feature string, tags []string) (string, error) {

	query := `
		SELECT tags.features_id FROM tags WHERE id_banner = $1 LIMIT 1
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

	old, err := stmt.Query(id)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	var oldFeature string
	if old.Next() {
		err := old.Scan(&oldFeature)
		if err != nil {
			return "", err
		}
	}
	if oldFeature == "" {
		return "NULL", err
	}
	if oldFeature == feature {
		fmt.Println(id, pq.Array(tags))
		query = `
			UPDATE tags SET tag_list = $2 WHERE id_banner = $1;
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

		_, err = stmt.Query(id, pq.Array(tags))
		if err != nil {
			return "", err
		}
		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {

			}
		}(stmt)
	} else {
		_, err = DeleterTags(db, id)
		if err != nil {
			return "", err
		}
		err = CreateNewFeatureStorage(db, feature)
		if err != nil {
			return "", err
		}
		err = CreateNewTagStorage(db, feature, id, tags)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func DeleterTags(db *sql.DB, item string) (string, error) {
	query := `WITH deleted AS (
    DELETE FROM tags 
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
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	rows, err := stmt.Query(item)
	if err != nil {
		return "", err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)
	var ans string
	if rows.Next() {
		err := rows.Scan(&ans)
		if err != nil {
			return "", err
		}
	}
	return ans, nil
}

func DeleterTagsPostponed(db *sql.DB, item []int) error {
	fmt.Println(item)
	query := `
    INSERT INTO delayed_deletions (table_name, id_item)
    SELECT 'tags', unnest($1::int[]);
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
