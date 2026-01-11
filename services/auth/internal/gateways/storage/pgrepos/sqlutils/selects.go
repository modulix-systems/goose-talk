package sqlutils

import (
	"fmt"
	"reflect"
	"strings"

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

			switch fieldType.Name() {
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

var UserSelect = strings.Join(getUserSelectCols(), ",")
