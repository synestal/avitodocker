package get_funcs

import (
	getsql "awesomeProject/internal/app/sqlDAO/get"
	help "awesomeProject/pkg/func"
	"database/sql"
	"sync"
)

func GetBannerByFilter(db *sql.DB, token, feature, limit, offset, tag string) (int, []getsql.FilteredBanner, error) {
	filteredBanner := make([]getsql.FilteredBanner, 0, 1)
	avaliable, adminState, err := getsql.GetAdminState(db, token)
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
		ids, err := getsql.GetBannerIdByFeature(db, feature, limit, offset)
		if err != nil {
			return 400, filteredBanner, err
		}
		err = getsql.GetBannerStorage(db, ids, &filteredBanner)
		if err != nil {
			return 400, filteredBanner, err
		}
	} else if feature == "" {
		ids, err := getsql.GetBannerIdByTag(db, tag, limit, offset)
		if err != nil {
			return 400, filteredBanner, err
		}
		err = getsql.GetBannerStorage(db, ids, &filteredBanner)
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
			idsFea, err = getsql.GetBannerIdByFeature(db, feature, limit, offset)
			if err != nil {

			}
		}()
		go func() {
			defer wg.Done()
			idsTag, err = getsql.GetBannerIdByTag(db, tag, limit, offset)
			if err != nil {

			}
		}()
		wg.Wait()

		ids := help.Intersection(idsFea, idsTag)
		err = getsql.GetBannerStorage(db, ids, &filteredBanner)
		if err != nil {
			return 400, filteredBanner, err
		}
	}
	return 200, filteredBanner, nil
}

func GetBannersHistory(db *sql.DB, token, id string) (int, []getsql.HistoryBanner, error) {
	filteredBanner := make([]getsql.HistoryBanner, 0, 1)
	avaliable, adminState, err := getsql.GetAdminState(db, token)
	if err != nil {
		return 400, filteredBanner, err
	}
	if avaliable == false {
		return 401, filteredBanner, err
	}
	if adminState == false {
		return 403, filteredBanner, err // Пользователь не имеет доступа
	}
	err = getsql.GetBannerHistoryStorage(db, id, &filteredBanner)
	if err != nil {
		return 400, filteredBanner, err
	}
	return 200, filteredBanner, nil
}
