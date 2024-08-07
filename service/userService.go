package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"go.uber.org/zap"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"

	"github.com/CCLooMi/sql-mak/mysql/mak"
)

type UserService struct {
	*dao.BaseDao
	db  *sql.DB
	log *zap.Logger
}

func NewUserService(db *sql.DB, log *zap.Logger) *UserService {
	return &UserService{BaseDao: dao.NewBaseDao(db), db: db, log: log}
}

func (dao *UserService) FindById(id *string) (*entity.User, error) {
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
		user.Id = new(string)
		*user.Id = utils.UUID()
	}
	if user.InsertedBy == nil {
		user.InsertedBy = user.Id
		user.UpdatedBy = user.Id
	}
	return dao.SaveOrUpdate(user)
}

func (dao *UserService) FindByUsernameAndPassword(username string, password string) (*entity.User, []entity.Role, map[string]bool) {
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

	var permissions []string
	sm = mysql.SELECT("rp.permission_id").
		FROM(entity.RolePermission{}, "rp").
		LEFT_JOIN(entity.RoleUser{}, "ru", "rp.role_id = ru.role_id").
		WHERE("ru.user_id = ?", user.Id)
	dao.FindBySM(sm, &permissions)
	var pm = make(map[string]bool)
	for _, v := range permissions {
		pm[v] = true
	}
	return &user, roles, pm
}

func (dao *UserService) CheckExist(e *entity.User) bool {
	sm := mysql.SELECT("COUNT(1)").
		FROM(entity.User{}, "e").
		WHERE("e.username = ?", e.Username).
		LIMIT(1)
	return dao.ExecuteSM(sm).Count() > 0
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

func (dao *UserService) DeleteUser(e *entity.User) []sql.Result {
	tx, err := dao.db.Begin()
	if err != nil {
		panic(err.Error())
	}
	dm := mysql.DELETE().FROM(entity.RoleUser{}).
		WHERE("user_id = ?", e.Id)
	dm2 := mysql.DELETE().FROM(entity.User{}).
		WHERE("id = ?", e.Id)
	rs := mysql.TxExecute(tx, dm, dm2)
	return rs
}
func (dao *UserService) FindRolesByUserId(userId string, page int, pageSize int, yes bool) map[string]interface{} {
	var roles []entity.Role
	count, err := dao.ByPage(&roles, page, pageSize, func(sm *mak.SQLSM) {
		sm.SELECT("DISTINCT r.*").
			FROM(entity.Role{}, "r").
			LEFT_JOIN(entity.RoleUser{}, "ru", "ru.role_id = r.id")
		if yes {
			sm.WHERE("ru.user_id = ?", userId)
		} else {
			sm.WHERE("(ISNULL(ru.user_id) OR ru.user_id <> ?)", userId)
		}
	})
	if err != nil {
		dao.log.Warn(err.Error())
	}
	return map[string]interface{}{
		"total": count,
		"data":  roles,
	}
}

func (dao *UserService) AddRole(e *entity.RoleUser) sql.Result {
	return dao.SaveOrUpdate(e)
}
func (dao *UserService) RemoveRole(e *entity.RoleUser) sql.Result {
	return dao.Delete(e)
}
