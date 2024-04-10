package postgres

import "database/sql"

func CreateTable(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS features (
    features TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS tags (
    id_banner INTEGER,
    tag_list INTEGER[],
    features TEXT REFERENCES features(features),
    PRIMARY KEY (id_banner)
);

CREATE TABLE IF NOT EXISTS banners_storage(
    id_banner  SERIAL,
    title_banner TEXT,
    text_banner TEXT,
    url_banner TEXT,
    banner_state BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id_banner)
);

CREATE TABLE IF NOT EXISTS delayed_deletions(
    id  SERIAL,
    table_name TEXT,
    id_item INTEGER,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS user_tokens(
    token_state  BOOLEAN,
    id INTEGER
);

CREATE TABLE IF NOT EXISTS history_banenrs(
    id SERIAL,
    id_banner  integer,
    title_banner TEXT,
    text_banner TEXT,
    url_banner TEXT,
    banner_state BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS history_features (
    id SERIAL,
    id_banner INTEGER,
    features_id INTEGER,
    tag_list INTEGER[],
    PRIMARY KEY (id)
);
		`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func CreateTrigger(db *sql.DB) error {
	query := `
CREATE OR REPLACE FUNCTION process_delayed_deletions()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.table_name = 'tags' OR NEW.table_name = 'banners_storage' THEN
        EXECUTE format('DELETE FROM %I WHERE id_banner = ANY(SELECT id_item FROM delayed_deletions WHERE table_name = %L)', NEW.table_name, NEW.table_name);
		DELETE FROM delayed_deletions WHERE table_name = NEW.table_name;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS before_update_or_insert_on_delayed_deletions ON delayed_deletions;

CREATE TRIGGER before_update_or_insert_on_delayed_deletions
AFTER INSERT ON delayed_deletions
FOR EACH ROW
EXECUTE FUNCTION process_delayed_deletions();
------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_history()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO history_features (id_banner, features_id, tag_list)
    VALUES (NEW.id_banner, NEW.features_id, NEW.tag_list);

    DELETE FROM history_features
    WHERE id_banner = NEW.id_banner
    AND id IN (
        SELECT id
        FROM history_features
        WHERE id_banner = NEW.id_banner
        ORDER BY id DESC
        OFFSET 4
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_history_trigger ON tags;

CREATE TRIGGER  update_history_trigger
AFTER UPDATE OF tag_list ON tags
FOR EACH ROW
EXECUTE FUNCTION update_history();

DROP TRIGGER IF EXISTS insert_history_trigger ON tags;

CREATE TRIGGER insert_history_trigger
AFTER INSERT ON tags
FOR EACH ROW
EXECUTE FUNCTION update_history();
------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_history_banner()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO history_banenrs (id_banner, title_banner, text_banner, url_banner, banner_state, created_at, updated_at)
    VALUES (NEW.id_banner, NEW.title_banner, NEW.text_banner, NEW.url_banner, NEW.banner_state, NEW.created_at, NEW.updated_at);

    DELETE FROM history_banenrs
    WHERE id_banner = NEW.id_banner
    AND id IN (
        SELECT id
        FROM history_banenrs
        WHERE id_banner = NEW.id_banner
        ORDER BY id DESC
        OFFSET 4
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_history_banner_trigger ON banners_storage;

CREATE TRIGGER update_history_banner_trigger
AFTER UPDATE OF title_banner ON banners_storage
FOR EACH ROW
EXECUTE FUNCTION update_history_banner();

DROP TRIGGER IF EXISTS insert_history_banner_trigger ON banners_storage;

CREATE TRIGGER insert_history_banner_trigger
AFTER INSERT ON banners_storage
FOR EACH ROW
EXECUTE FUNCTION update_history_banner();
------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_updated_at_trigger ON banners_storage;

CREATE TRIGGER update_updated_at_trigger
BEFORE UPDATE ON banners_storage
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
		`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
