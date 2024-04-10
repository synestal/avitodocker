package post_funcs

import (
	getsql "awesomeProject/internal/app/sqlDAO/get"
	postsql "awesomeProject/internal/app/sqlDAO/post"
	"database/sql"
)

func CreateNewBanner(db *sql.DB, token, feature, active string, content, tags []string) (int, *postsql.BannerId, error) {
	avaliable, adminState, err := getsql.GetAdminState(db, token)
	if err != nil {
		return 500, nil, err
	}
	if avaliable == false {
		return 401, nil, err
	}
	if adminState == false {
		return 403, nil, err
	}

	bannerId, err := postsql.CreateNemBannerStorage(db, active, content)
	if err != nil {
		return 500, nil, err
	}
	err = postsql.CreateNewFeatureStorage(db, feature)
	if err != nil {
		return 500, nil, err
	}
	err = postsql.CreateNewTagStorage(db, feature, bannerId.ID, tags)
	if err != nil {
		return 500, nil, err
	}

	return 201, bannerId, nil
}

func ChangeBanner(db *sql.DB, token, bannerid, feature, active string, content, tags []string) (int, error) {
	avaliable, adminState, err := getsql.GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}

	ans, err := postsql.UpdateBannersStorage(db, bannerid, active, content)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	ans, err = postsql.UpdateFeatureTagStorage(db, bannerid, feature, tags)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	return 200, nil
}

func DeleteBanner(db *sql.DB, token, bannerid string) (int, error) {
	avaliable, adminState, err := getsql.GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}

	ans, err := postsql.DeleterTags(db, bannerid)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	ans, err = postsql.DeleterBanners(db, bannerid)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	return 204, nil
}

func DeleteBannerByFeatureOrTag(db *sql.DB, token, feature, limit, offset, tag string) (int, error) {
	avaliable, adminState, err := getsql.GetAdminState(db, token)
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
		ids, err = getsql.GetBannerIdByFeature(db, feature, limit, offset)
		if err != nil {
			return 400, err
		}
	} else {
		ids, err = getsql.GetBannerIdByTag(db, tag, limit, offset)
		if err != nil {
			return 400, err
		}
	}

	tagsErrChan := make(chan error)
	bannersErrChan := make(chan error)
	go func() {
		tagsErrChan <- postsql.DeleterTagsPostponed(db, ids)
	}()
	go func() {
		bannersErrChan <- postsql.DeleterBannersPostponed(db, ids)
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

func ChangeBannersHistory(db *sql.DB, token, number, id string) (int, error) {
	avaliable, adminState, err := getsql.GetAdminState(db, token)
	if err != nil {
		return 500, err
	}
	if avaliable == false {
		return 401, err
	}
	if adminState == false {
		return 403, err
	}

	ans, err := postsql.ChangeHistoryBannersStorage(db, number, id)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	ans, err = postsql.ChangeHistoryFeatureTagStorage(db, number, id)
	if err != nil {
		return 500, err
	}
	if ans == "NULL" {
		return 404, nil
	}
	return 200, nil
}
