package sqlutils

import (
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/entity"
)

func RowToUser(row pgx.CollectableRow) (entity.User, error) {
	var user entity.User
	userData, err := pgx.RowToMap(row)
	if err != nil {
		return user, err
	}
	// group related entities flat data into map
	userStructJsonData := make(map[string]any)
	for key, value := range userData {
		keyParts := strings.Split(key, ".")
		var prefix, fieldName string = keyParts[0], keyParts[1]
		switch prefix {
		case userColPrefix:
			userStructJsonData[fieldName] = value
		case twoFaColPrefix:
			twoFaData, ok := userStructJsonData[twoFaColPrefix].(map[string]any)
			if !ok {
				twoFaData = make(map[string]any)
				userStructJsonData[twoFaColPrefix] = twoFaData
			}
			twoFaData[fieldName] = value
		}
	}
	userJson, err := json.Marshal(userStructJsonData)
	if err != nil {
		return user, err
	}
	if err := json.Unmarshal(userJson, &user); err != nil {
		return user, err
	}
	return user, nil
}
