package loaders

import (
	"database/sql"
	"log/slog"
	"strings"
	"tracker/logging"
	"tracker/types"
)

func UserAccounts(db *sql.DB) (*[]types.Account, error) {
	log := logging.Get()
	rows, err := db.Query("SELECT id,name,institution,tags from accounts order by CAST(id as decimal)")
	if err != nil {
		log.Error("failed to all accounts for user", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	accounts := make([]types.Account, 0)
	for rows.Next() {
		var account types.Account
		var tagsStr sql.NullString
		_ = rows.Scan(&account.Id, &account.Name, &account.Institution, &tagsStr)
		if tagsStr.Valid && tagsStr.String != "" {
			account.Tags = strings.Split(tagsStr.String, ",")
		} else {
			account.Tags = []string{}
		}
		accounts = append(accounts, account)
	}

	return &accounts, nil
}
