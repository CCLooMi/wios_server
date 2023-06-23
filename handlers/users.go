package handlers

import (
	"wios_server/dao"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

// 用户列表查询路由
func GetUserListHandler(db *gorm.DB) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		// 获取请求参数
		pageNumber, _ := ctx.URLParamInt("pageNumber")
		pageSize, _ := ctx.URLParamInt("pageSize")

		// 查询用户列表
		userDao := dao.NewUserDao(db)
		count, users, err := userDao.FindByPage(pageNumber, pageSize)
		if err != nil {
			panic(err)
		}
		result := map[string]interface{}{
			"count": count,
			"users": users,
		}
		// 返回用户列表
		ctx.JSON(result)
	}
}
