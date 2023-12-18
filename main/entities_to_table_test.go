package main

import (
	"fmt"
	"github.com/CCLooMi/sql-mak/utils"
	"strings"
	"testing"
	"wios_server/conf"
	"wios_server/entity"
)

func TestEntitiesToTable(t *testing.T) {
	sql := createTable(
		&entity.Menu{},
		&entity.Org{},
		&entity.OrgUser{},
		&entity.Permission{},
		&entity.Role{},
		&entity.RoleMenu{},
		&entity.RolePermission{},
		&entity.RoleUser{},
		&entity.Upload{},
		&entity.User{},
		&entity.Account{},
		&entity.Category{},
		&entity.Comment{},
		&entity.PurchasedWpp{},
		&entity.Wpp{},
		&entity.Api{})

	defer conf.Db.Close()

	r, err := conf.Db.Exec(sql)
	if err != nil {
		t.Error(err)
	}
	t.Log(r.RowsAffected())
}
func createTable(entities ...interface{}) string {
	var sqlStatements []string

	for _, entity := range entities {
		entityInfo := utils.GetEntityInfo(entity)

		var columns []string
		for i, columnName := range entityInfo.Columns {
			tag := entityInfo.Tags[i]
			columnType := strings.TrimSpace(strings.Split(tag.Get("orm"), ";")[0])

			column := fmt.Sprintf("`%s` %s", columnName, strings.ToUpper(columnType))
			columns = append(columns, column)
		}

		primaryKey := entityInfo.PrimaryKey
		primaryKeyConstraint := fmt.Sprintf("PRIMARY KEY (%s)", primaryKey)

		sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\nCREATE TABLE `%s` (\n\t%s,\n\t%s\n);",
			entityInfo.TableName,
			entityInfo.TableName,
			strings.Join(columns, ",\n\t"),
			primaryKeyConstraint)
		fmt.Println(sql)
		sqlStatements = append(sqlStatements, sql)
	}

	return strings.Join(sqlStatements, "\n\n")
}
