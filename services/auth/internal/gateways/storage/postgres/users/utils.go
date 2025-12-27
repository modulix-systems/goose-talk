package users_repo

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/utils"
)

const userColPrefix = "user"
const twoFaColPrefix = "two_factor_auth"

func formatColName(prefix string, name string) string {
	colName := utils.ToSnakeCase(name)
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, colName)
	}
	return colName
}

func getSelectColNameWithAlias(tableName string, name string) string {
	return fmt.Sprintf(
		`%s as "%s"`, formatColName(fmt.Sprintf(`"%s"`, tableName), name),
		formatColName(strings.ReplaceAll(tableName, "\"", ""), name),
	)
}

func getUserSelectCols() []string {
	userType := reflect.TypeFor[entity.User]()
	colsList := make([]string, 0, userType.NumField()+1)
	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		if !utils.IsScalarType(field.Type) {
		fieldType := field.Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}

			switch (fieldType.Name()) {
			case "Time":
			case "TwoFactorAuth":
				for i := 0; i < fieldType.NumField(); i++ {
					colsList = append(colsList, getSelectColNameWithAlias(twoFaColPrefix, fieldType.Field(i).Name))
				}
				continue
			default:
				continue
			}
		}
		colsList = append(colsList, getSelectColNameWithAlias(userColPrefix, field.Name))
	}
	return colsList
}

var userSelect = strings.Join(getUserSelectCols(), ",")

func parseUserFromRow(row pgx.CollectableRow) (entity.User, error) {
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