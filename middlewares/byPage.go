package middlewares

import (
	"github.com/gin-gonic/gin"
	"wios_server/handlers/msg"
)

type Page struct {
	PageNumber int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	Opts       map[string]interface{} `json:"opts"`
}

func ByPage(c *gin.Context, f func(page *Page) (int64, any, error)) {
	var page Page
	if err := c.ShouldBindJSON(&page); err != nil {
		page = Page{
			PageNumber: 1,
			PageSize:   20,
			Opts:       map[string]interface{}{},
		}
	}
	if page.PageNumber <= 0 {
		page.PageNumber = 0
	} else {
		page.PageNumber = page.PageNumber - 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	count, data, err := f(&page)
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	result := map[string]interface{}{
		"total": count,
		"data":  data,
	}
	msg.Ok(c, result)
}

func ByPageMap(reqBody map[string]interface{}, c *gin.Context, f func(page *Page) (int64, any, error)) {
	var page Page
	pageNumber, ok := reqBody["page"].(float64)
	if !ok {
		pageNumber = 0
	}
	pageSize, ok := reqBody["pageSize"].(float64)
	if !ok {
		pageSize = 20
	}
	opts, ok := reqBody["opts"].(map[string]interface{})
	if !ok {
		opts = map[string]interface{}{}
	}
	page = Page{
		PageNumber: int(pageNumber),
		PageSize:   int(pageSize),
		Opts:       opts,
	}
	if page.PageNumber <= 0 {
		page.PageNumber = 0
	} else {
		page.PageNumber = page.PageNumber - 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	count, data, err := f(&page)
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	result := map[string]interface{}{
		"total": count,
		"data":  data,
	}
	msg.Ok(c, result)
}
