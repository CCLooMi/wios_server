package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"

	"github.com/CCLooMi/sql-mak/mysql/mak"
)

type UserService struct {
	*dao.BaseDao
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{BaseDao: dao.NewBaseDao(db)}
}

func (dao *UserService) FindById(id uint) (*entity.User, error) {
	var user entity.User
	dao.ById(id, &user)
	return &user, nil
}

func (dao *UserService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.User, error) {
	var users []entity.User
	count, err := dao.ByPage(&users, pageNumber, pageSize, fn)
	if err != nil {
		return 0, users, err
	}
	return count, users, nil
}
func (dao *UserService) SaveUpdate(user *entity.User) sql.Result {
	if user.Id == nil {
		*user.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(user)
}

func (dao *UserService) FindByUsernameAndPassword(username string, password string) (*entity.User, *[]entity.Role, *[]entity.Permission) {
	var user entity.User
	sm := mysql.SELECT("*").
		FROM(user, "u").
		WHERE("u.username = ?", username).
		AND("u.password = SHA2(CONCAT(u.username,?,u.seed),256)", password).
		LIMIT(1)
	dao.FindBySM(sm, &user)
	if user.Id == nil {
		return nil, nil, nil
	}
	var roles []entity.Role
	sm = mysql.SELECT("r.*").
		FROM(entity.RoleUser{}, "ur").
		LEFT_JOIN(entity.Role{}, "r", "ur.role_id = r.id").
		WHERE("ur.user_id = ?", user.Id)
	dao.FindBySM(sm, &roles)

	var permissions []entity.Permission
	sm = mysql.SELECT("p.*").
		FROM(entity.RolePermission{}, "rp").
		LEFT_JOIN(entity.Permission{}, "p", "rp.permission_id = p.id").
		LEFT_JOIN(entity.RoleUser{}, "ru", "rp.role_id = ru.role_id").
		WHERE("ru.user_id = ?", user.Id)
	dao.FindBySM(sm, &permissions)
	return &user, &roles, &permissions
}

func (dao *UserService) CheckExist(e *entity.User) bool {
	var user entity.User
	sm := mysql.SELECT("*").
		FROM(user, "e").
		WHERE("e.username = ?", e.Username).
		LIMIT(1)
	dao.FindBySM(sm, &user)
	return user.Id != nil
}

func (dao *UserService) FindMenusByUser(user *entity.User) []entity.Menu {
	var menus []entity.Menu
	if user.Username == "root" {
		sm := mysql.SELECT("m.*").
			FROM(entity.Menu{}, "m")
		dao.FindBySM(sm, &menus)
		return menus
	}
	sm := mysql.SELECT("DISTINCT m.*").
		FROM(entity.RoleMenu{}, "rm").
		LEFT_JOIN(entity.Menu{}, "m", "rm.menu_id = m.id").
		LEFT_JOIN(entity.RoleUser{}, "ru", "rm.role_id = ru.role_id").
		WHERE("ru.user_id = ?", user.Id)
	dao.FindBySM(sm, &menus)
	return menus
}
