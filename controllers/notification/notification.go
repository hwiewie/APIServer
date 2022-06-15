package notification

import (
	"encoding/json"
	"strconv"

	"github.com/hwiewie/APIServer/controllers/base"
	"github.com/hwiewie/APIServer/models"
	"github.com/hwiewie/APIServer/util/logs"
)

type NotificationController struct {
	base.APIController
}

func (c *NotificationController) URLMapping() {
	c.Mapping("List", c.List)
	c.Mapping("Create", c.Create)
	c.Mapping("Publish", c.Publish)
	c.Mapping("Subscribe", c.Subscribe)
	c.Mapping("Read", c.Read)

}

func (c *NotificationController) Prepare() {
	c.APIController.Prepare()
	_, method := c.GetControllerAndAction()
	switch method {
	case "List", "Create", "Publish":
		if !c.User.Admin {
			// 只有管理員才能查看、創建和廣播消息
			c.AbortForbidden("operation need admin permission.")
		}
	default:
	}
}

// 列出
// @router / [get]
func (c *NotificationController) List() {
	param := c.BuildQueryParam()
	notifications := []models.Notification{}
	total, err := models.GetTotal(new(models.Notification), param)
	if err != nil {
		logs.Error("get total count by param (%v) error. %v", param, err)
		c.HandleError(err)
		return
	}

	param.Relate = "all"
	err = models.GetAll(new(models.Notification), &notifications, param)
	if err != nil {
		logs.Error("list by param (%v) error. %v", param, err)
		c.HandleError(err)
		return
	}
	c.Success(param.NewPage(total, notifications))
}

// 創建
// @router / [post]
func (c *NotificationController) Create() {
	var notify models.Notification
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &notify)
	if err != nil {
		logs.Error("get body error. %v", err)
		c.AbortBadRequestFormat("Notification")
	}

	notify.FromUser = c.User
	_, err = models.NotificationModel.Add(&notify)

	if err != nil {
		logs.Error("create error.%v", err.Error())
		c.HandleError(err)
		return
	}
	c.Success(notify)
}

// 發布
// @router / [put]
func (c *NotificationController) Publish() {
	id := c.Input().Get("id")
	id64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.HandleError(err)
		return
	}
	err = models.NotificationLogModel.AddToAllUser(id64)
	if err != nil {
		c.HandleError(err)
		return
	}
	_ = models.NotificationModel.UpdateById(&models.Notification{Id: id64, IsPublished: true})
	c.Success(nil)
}

// 用戶訂閱
// @router /subscribe [get]
func (c *NotificationController) Subscribe() {
	param := c.BuildQueryParam()
	param.Query["user_id"] = c.User.Id
	IsReaded := c.Input().Get("is_readed")
	if IsReaded != "" {
		r, err := strconv.ParseBool(IsReaded)
		if err != nil {
			c.HandleError(err)
			return
		}
		param.Query["is_readed"] = r
	}
	param.Relate = "all"
	notificationlogs := []models.NotificationLog{}

	err := models.GetAll(new(models.NotificationLog), &notificationlogs, param)
	if err != nil {
		c.HandleError(err)
		return
	}

	c.Success(notificationlogs)
}

// 用戶已讀
// @router /subscribe [put]
func (c *NotificationController) Read() {
	id := c.Input().Get("id")
	id64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.HandleError(err)
		return
	}
	l := models.NotificationLog{
		UserId:   c.User.Id,
		Id:       id64,
		IsReaded: true,
	}

	err = models.NotificationLogModel.UpdateByUserId(&l)
	if err != nil {
		c.HandleError(err)
		return
	}
	c.Success(nil)
}
