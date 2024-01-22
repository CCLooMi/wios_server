package main

import (
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	baseEntity "github.com/CCLooMi/sql-mak/mysql/entity"
	"github.com/CCLooMi/sql-mak/utils"
	"strings"
	"testing"
	"wios_server/conf"
	"wios_server/entity"
	utils2 "wios_server/utils"
)

var tables = []interface{}{
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
	&entity.Api{},
	&entity.Config{},
}

func TestEntitiesToTable(t *testing.T) {
	sqls := createTable(tables...)
	//execute sql
	defer conf.Db.Close()
	for _, sql := range sqls {
		_, err := conf.Db.Exec(sql)
		if err != nil {
			t.Error(sql, err)
		}
		t.Log("execute sql success")
	}
}

func TestEntitiesToTableIfNotExist(t *testing.T) {
	sqls := createIfNotExistTable(tables...)
	//execute sql
	defer conf.Db.Close()
	for _, sql := range sqls {
		_, err := conf.Db.Exec(sql)
		if err != nil {
			t.Error(sql, err)
		}
		t.Log("execute sql success")
	}
}

func TestCreateRootUser(t *testing.T) {
	defer conf.Db.Close()
	Id := "3d81bff4b8cc11ee82370242ac120002"
	seed := utils2.RandomBytes(8)
	pass := utils2.SHA256("root", "apple", seed)
	im := mysql.INSERT_INTO(entity.User{
		IdEntity: baseEntity.IdEntity{Id: &Id},
		Username: "root",
		Nickname: "Root",
		Password: pass,
		Seed:     seed,
	}).ON_DUPLICATE_KEY_UPDATE().
		SET("username=?", "root").
		SET("nickname=?", "Root").
		SET("password=?", pass).
		SET("seed=?", seed)
	r := im.Execute(conf.Db).Update()
	rc, _ := r.RowsAffected()
	t.Log("Create root user result:", rc)
}
func createTable(entities ...interface{}) []string {
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

		sqlStatements = append(sqlStatements,
			fmt.Sprintf("DROP TABLE IF EXISTS `%s`", entityInfo.TableName))
		sqlStatements = append(sqlStatements, fmt.Sprintf("CREATE TABLE `%s` (\n\t%s,\n\t%s\n)",
			entityInfo.TableName,
			strings.Join(columns, ",\n\t"),
			primaryKeyConstraint))
	}
	return sqlStatements
}
func createIfNotExistTable(entities ...interface{}) []string {
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

		sqlStatements = append(sqlStatements, fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n\t%s,\n\t%s\n)",
			entityInfo.TableName,
			strings.Join(columns, ",\n\t"),
			primaryKeyConstraint))
	}
	return sqlStatements
}
