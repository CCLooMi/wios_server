package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/handlers/msg"
)

type ApiController struct {
	db *sql.DB
}

func (ctrl *ApiController) execute(c *gin.Context) {
	var reqBody ReqBody
	if err := c.BindJSON(&reqBody); err != nil {
		msg.Error(c, err)
		return
	}

	sm := mak.NewSQLSM()
	sm.SELECT("script").
		FROM("sys_api", "a").
		WHERE("a.id = ?", reqBody.ID).
		LIMIT(1)
	stmp, err := ctrl.db.Prepare(sm.Sql())
	if err != nil {
		msg.Error(c, err)
		return
	}
	rows, err := stmp.Query(sm.Args()...)
	if err != nil {
		msg.Error(c, err)
		return
	}
	for rows.Next() {
		var data string
		err := rows.Scan(&data)
		if err != nil {
			msg.Error(c, err)
			return
		}
		msg.Ok(c, data)
		return
	}
	msg.Ok(c, nil)
}

type ReqBody struct {
	ID   string `json:"id"`
	Args []any  `json:"args"`
}

func NewApiController(app *gin.Engine, db *sql.DB) *ApiController {
	ctrl := &ApiController{db: db}
	group := app.Group("/api")
	group.POST("/execute", ctrl.execute)
	return ctrl
}
