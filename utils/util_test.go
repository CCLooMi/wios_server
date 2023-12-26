package utils

import (
	"fmt"
	"github.com/CCLooMi/sql-mak/utils"
	"testing"
	"wios_server/handlers/beans"
)

func TestEntityInfo(t *testing.T) {
	entityInfo := utils.GetEntityInfo(beans.MenuWithChecked{})
	fmt.Println(entityInfo.TableName)
}
