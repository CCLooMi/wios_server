package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"io"
	"net/http"
	"strings"
	"time"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/js"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type ApiController struct {
	db         *sql.DB
	apiService *service.ApiService
}

func NewApiController(app *gin.Engine) *ApiController {
	ctrl := &ApiController{db: conf.Db, apiService: service.NewApiService(conf.Db)}
	group := app.Group("/api")
	hds := []middlewares.Auth{
		{Method: "GET", Group: "/api", Path: "/stopVmById", Auth: "api.stopVmById", Handler: ctrl.stopVmById},
		{Method: "POST", Group: "/api", Path: "/vms", Auth: "api.vms", Handler: ctrl.vms},
		{Method: "POST", Group: "/api", Path: "/execute", Auth: "api.execute", Handler: ctrl.execute},
		{Method: "POST", Group: "/api", Path: "/executeById", Auth: "#", Handler: ctrl.executeById, AuthCheck: middlewares.ScriptApiAuthCheck},
		{Method: "POST", Group: "/api", Path: "/byPage", Auth: "api.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/api", Path: "/saveUpdate", Auth: "api.saveUpdate", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/api", Path: "/saveUpdates", Auth: "api.saveUpdates", Handler: ctrl.saveUpdates},
		{Method: "POST", Group: "/api", Path: "/delete", Auth: "api.delete", Handler: ctrl.delete},
		{Method: "GET", Group: "/api", Path: "/backup", Auth: "api.backup", Handler: ctrl.backup},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

var halt = errors.New("Stahp")

func closeChannel(c chan func()) {
	select {
	case c <- func() {}:
		close(c)
	default:
	}
}

type vmMeta struct {
	Title *string       `json:"title"`
	User  *entity.User  `json:"user"`
	Exit  func() string `json:"-"`
}

func (vm *vmMeta) SetTitle(t *string) {
	vm.Title = t
}

var vmMap = make(map[string]*vmMeta)

func runUnsafe(unsafe string, title *string, c *gin.Context, args interface{}, reqBody map[string]interface{}) {
	ui, ok := c.Get("userInfo")
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}
	userInfo, ok := ui.(*middlewares.UserInfo)
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == halt {
				msg.Error(c, "Stopping after: "+duration.String())
				return
			}
			// Something else happened, repanic!
			panic(caught)
		}
	}()
	vmId := utils.UUID()
	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	var exit = func() string {
		vm.Interrupt <- func() {
			panic(halt)
		}
		delete(vmMap, vmId)
		closeChannel(vm.Interrupt)
		return "vm[" + vmId + "] exited."
	}
	var self = vmMeta{Title: title, User: userInfo.User, Exit: exit}
	vmMap[vmId] = &self
	defer closeChannel(vm.Interrupt)
	defer delete(vmMap, vmId)
	vm.Set("ctx", c)
	vm.Set("reqBody", reqBody)
	vm.Set("msgOk", func(data any) {
		msg.Ok(c, data)
	})
	vm.Set("msgError", func(err any) {
		msg.Error(c, err)
	})
	vm.Set("msgOks", func(data ...any) {
		msg.Oks(c, data...)
	})
	vm.Set("byPage", func(f func(sm *mak.SQLSM, opts interface{})) {
		middlewares.ByPageMap(reqBody, c, func(page *middlewares.Page) (int64, any, error) {
			if page.PageNumber < 0 {
				page.PageNumber = 0
			} else {
				page.PageNumber -= 1
			}
			sm := mak.NewSQLSM()
			f(sm, page.Opts)
			sm.LIMIT(page.PageNumber*page.PageSize, page.PageSize)
			out := sm.Execute(conf.Db).GetResultAsMapList()
			if page.PageNumber == 0 {
				return sm.Execute(conf.Db).Count(), out, nil
			}
			return -1, out, nil
		})
	})
	vm.Set("fetch", func(url string, opts ...interface{}) map[string]interface{} {
		result, err := fetch(url, opts...)
		if err != nil {
			msg.Error(c, err.Error())
			return nil
		}
		return result
	})
	vm.Set("$", func(str string) *goquery.Document {
		result, error := goquery.NewDocumentFromReader(strings.NewReader(str))
		if error != nil {
			msg.Error(c, error.Error())
			return nil
		}
		return result
	})
	vm.Set("userInfo", userInfo)
	vm.Set("args", args)
	vm.Set("exit", exit)
	vm.Set("self", &self)
	js.ApplyExportsTo(vm)
	result, err := vm.Run(unsafe)
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	if result.IsBoolean() {
		b, _ := result.ToBoolean()
		if !b {
			return
		}
	}
	msg.Ok(c, result)
	return
}
func getStrDecode(name string) *encoding.Decoder {
	switch strings.ToUpper(name) {
	case "UTF8", "UTF-8":
		return unicode.UTF8.NewDecoder()
	case "GBK", "GB2312":
		return simplifiedchinese.GBK.NewDecoder()
	case "GB18030":
		return simplifiedchinese.GB18030.NewDecoder()
	case "BIG5":
		return traditionalchinese.Big5.NewDecoder()
	case "ISO-8859-1":
		return charmap.ISO8859_1.NewDecoder()
	case "ISO-8859-2":
		return charmap.ISO8859_2.NewDecoder()
	case "ISO-8859-3":
		return charmap.ISO8859_3.NewDecoder()
	case "ISO-8859-4":
		return charmap.ISO8859_4.NewDecoder()
	case "ISO-8859-5":
		return charmap.ISO8859_5.NewDecoder()
	case "ISO-8859-6":
		return charmap.ISO8859_6.NewDecoder()
	case "ISO-8859-7":
		return charmap.ISO8859_7.NewDecoder()
	case "ISO-8859-8":
		return charmap.ISO8859_8.NewDecoder()
	case "ISO-8859-9":
		return charmap.ISO8859_9.NewDecoder()
	case "ISO-8859-10":
		return charmap.ISO8859_10.NewDecoder()
	case "WINDOWS-1250":
		return charmap.Windows1250.NewDecoder()
	case "WINDOWS-1251":
		return charmap.Windows1251.NewDecoder()
	case "WINDOWS-1252":
		return charmap.Windows1252.NewDecoder()
	case "WINDOWS-1253":
		return charmap.Windows1253.NewDecoder()
	case "WINDOWS-1254":
		return charmap.Windows1254.NewDecoder()
	case "WINDOWS-1255":
		return charmap.Windows1255.NewDecoder()
	case "WINDOWS-1256":
		return charmap.Windows1256.NewDecoder()
	case "WINDOWS-1257":
		return charmap.Windows1257.NewDecoder()
	case "WINDOWS-1258":
		return charmap.Windows1258.NewDecoder()
	case "KOI8-R":
		return korean.EUCKR.NewDecoder()
	case "EUC-JP":
		return japanese.EUCJP.NewDecoder()
	case "ISO-2022-JP":
		return japanese.ISO2022JP.NewDecoder()
	case "UTF-16", "UTF-16BE":
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	case "UTF-16LE":
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	}
	return unicode.UTF8.NewDecoder()
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

func fetch(url string, o ...interface{}) (map[string]interface{}, error) {
	var opts map[string]interface{}
	if len(o) > 0 {
		opts = o[0].(map[string]interface{})
	} else {
		opts = map[string]interface{}{}
	}
	method, ok := opts["method"].(string)
	if !ok {
		method = "GET"
	}
	var body string
	if bodyInterface, ok := opts["body"]; ok {
		body = bodyInterface.(string)
	}
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	headers, ok := opts["headers"].(map[string]interface{})
	if ok {
		for k, v := range headers {
			req.Header.Set(k, v.(string))
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	charset, ok := opts["charset"].(string)
	if !ok {
		charset = "UTF8"
	}
	decode := getStrDecode(charset)
	rspBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dc, err := decode.Bytes(rspBody)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"response":   string(dc),
		"request":    resp.Request,
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
		"header":     resp.Header,
		"cookies": func() []*http.Cookie {
			return resp.Cookies()
		},
	}, nil
}

func (ctrl *ApiController) vms(c *gin.Context) {
	//msg.Ok(c, vmMap)
	data := make([]interface{}, 0)
	for k, v := range vmMap {
		data = append(data, map[string]interface{}{
			"id":    k,
			"title": v.Title,
			"user":  v.User.Nickname,
		})
	}
	msg.Ok(c, map[string]interface{}{
		"data":  data,
		"total": len(data),
	})
}
func (ctrl *ApiController) stopVmById(c *gin.Context) {
	vmId := c.Query("id")
	vm := vmMap[vmId]
	if vm == nil {
		msg.Error(c, "vm not found")
		return
	}
	msg.Ok(c, vm.Exit())
}
func (ctrl *ApiController) execute(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.BindJSON(&reqBody); err != nil {
		msg.Error(c, err)
		return
	}
	id, ok := reqBody["id"].(string)
	if !ok {
		id = utils.UUID()
	}
	script, ok := reqBody["script"].(string)
	if !ok {
		msg.Ok(c, "")
		return
	}
	args, ok := reqBody["args"]
	if !ok {
		args = map[string]interface{}{}
	}
	runUnsafe(script, &id, c, args, reqBody)
}
func (ctrl *ApiController) executeById(c *gin.Context) {
	apiInfo := c.MustGet(middlewares.ApiInfoKey).(*middlewares.ApiInfo)
	api := apiInfo.Api
	runUnsafe(*api.Script, api.Desc, c, apiInfo.Args, apiInfo.ReqBody)
}
func (ctrl *ApiController) byPage(c *gin.Context) {
	middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.apiService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Api{}, "a")
			q := page.Opts["q"]
			if q != nil && q != "" {
				lik := "%" + q.(string) + "%"
				sm.AND("(a.id = ? OR a.desc LIKE ? OR a.category LIKE ?)", q, lik, lik)
			}
		})
	})
}

func (ctrl *ApiController) saveUpdate(c *gin.Context) {
	var api entity.Api
	if err := c.ShouldBindJSON(&api); err != nil {
		msg.Error(c, err)
		return
	}
	userInfo := c.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	api.UpdatedBy = userInfo.User.Id
	if api.InsertedBy == nil {
		api.InsertedBy = userInfo.User.Id
	}
	if api.UpdatedAt != nil {
		api.UpdatedAt = nil
	}
	var rs = ctrl.apiService.SaveUpdate(&api)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err)
		return
	}
	msg.Ok(c, &api)
}
func (ctrl *ApiController) saveUpdates(c *gin.Context) {
	var apis []entity.Api
	if err := c.ShouldBindJSON(&apis); err != nil {
		msg.Error(c, err)
		return
	}
	userInfo, ok := c.Get("userInfo")
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}
	userId := userInfo.(*middlewares.UserInfo).User.Id
	for i := 0; i < len(apis); i++ {
		apis[i].UpdatedBy = userId
		if apis[i].InsertedBy == nil {
			apis[i].InsertedBy = userId
		}
	}
	var rs = ctrl.apiService.SaveUpdates(apis)
	for _, r := range rs {
		_, err := r.RowsAffected()
		if err != nil {
			msg.Error(c, err)
			return
		}
	}
	msg.Ok(c, &apis)
}

func (ctrl *ApiController) delete(c *gin.Context) {
	var api entity.Api
	if err := c.ShouldBindJSON(&api); err != nil {
		msg.Error(c, err)
		return
	}
	var rs = ctrl.apiService.Delete(&api)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err)
		return
	}
	if affected > 0 {
		msg.Ok(c, &api)
		return
	}
	msg.Error(c, "delete failed")
}

func (ctrl *ApiController) backup(c *gin.Context) {
	err := ctrl.apiService.Backup()
	if err != nil {
		msg.Error(c, err)
		return
	}
	msg.Ok(c, "ok")
}
