package base

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"

	"github.com/hwiewie/APIServer/bus"
	"github.com/hwiewie/APIServer/bus/message"
	"github.com/hwiewie/APIServer/models"
	"github.com/hwiewie/APIServer/models/response"
	"github.com/hwiewie/APIServer/util/hack"
)

type APIKeyController struct {
	beego.Controller

	APIKey  *models.APIKey
	Action  string
	Success response.Success
	Failure response.Failure
}

/**
 * 通過 apikey 參數判斷調用權限
 * apikey 類型：全局apikey（管理員可用）、命名空間級別的 apikey 和項目級別的 apikey（app 內部可用）
 **/
func (c *APIKeyController) Prepare() {
	c.Controller.Prepare()

	token := c.GetString("apikey")
	if token == "" {
		c.AddErrorAndResponse("No parameter named apikey in url query!", http.StatusForbidden)
		return
	}
	key, err := models.ApiKeyModel.GetByToken(token)
	// TODO 考慮統一處理 DB 錯誤
	if err == orm.ErrNoRows {
		c.AddErrorAndResponse("Invalid apikey parameter!", http.StatusForbidden)
		return
	} else if err != nil {
		c.AddErrorAndResponse("DB Connection Error!", http.StatusInternalServerError)
		return
	}
	if key.Deleted {
		c.AddErrorAndResponse("Invalid apikey parameter: deleted!", http.StatusForbidden)
		return
	}
	if key.ExpireIn != 0 && time.Now().After(key.CreateTime.Add(time.Second*time.Duration(key.ExpireIn))) {
		c.AddErrorAndResponse("Invalid apikey parameter: out of date!", http.StatusForbidden)
		return
	}
	_, c.Action = c.GetControllerAndAction()
	c.APIKey = key
}

func (c *APIKeyController) publishRequestMessage(code int, data interface{}) {
	var err error
	controller, _ := c.GetControllerAndAction()
	var body []byte
	switch val := data.(type) {
	case string:
		body = hack.Slice(val)
	default:
		body, _ = json.Marshal(data)
	}
	u := "[APIKey]"
	if c.APIKey != nil {
		u = c.APIKey.String()
	}
	messageData, err := json.Marshal(message.RequestMessageData{
		URI:            c.Ctx.Input.URI(),
		Controller:     controller,
		Method:         c.Action,
		User:           u,
		IP:             c.Ctx.Input.IP(),
		ResponseStatus: code,
		ResponseBody:   body,
	})
	if err != nil {
		logs.Error(err)
	} else {
		msg := message.Message{
			Type: message.TypeRequest,
			Data: json.RawMessage(messageData),
		}
		if err := bus.Notify(msg); err != nil {
			logs.Error(err)
		}
	}
}

// 用於負責 get 數據的接口，當 error 列表不為空的時候，返回 error 列表
// 當 參數為 nil 的時候，返回 "200"
func (c *APIKeyController) HandleResponse(data interface{}) {
	if len(c.Failure.Body.Errors) > 0 {
		c.Failure.Body.Code = http.StatusInternalServerError
		c.HandleByCode(http.StatusInternalServerError)
		return
	}
	if data == nil {
		c.Success.Body.Code = http.StatusOK
		data = c.Success.Body
	}
	c.publishRequestMessage(http.StatusOK, data)
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = data
	c.ServeJSON()
}

func (c *APIKeyController) HandleByCode(code int) {
	c.Ctx.Output.SetStatus(code)
	// gateway 處驗證不通過的狀態碼為 403
	if code < 400 {
		c.Success.Body.Code = code
		c.publishRequestMessage(code, c.Success)
		c.Data["json"] = c.Success.Body

	} else {
		c.Failure.Body.Code = code
		c.publishRequestMessage(code, c.Failure)
		c.Data["json"] = c.Failure.Body
	}
	c.ServeJSON()
}

func (c *APIKeyController) AddError(err string) {
	c.Failure.Body.Errors = append(c.Failure.Body.Errors, err)
}

func (c *APIKeyController) AddErrorAndResponse(err string, code int) {
	if code < 400 {
		panic("Not Error Code!")
	}
	if len(err) == 0 {
		err = http.StatusText(code)
	}
	c.AddError(err)
	c.HandleByCode(code)
	panic(err)
}
