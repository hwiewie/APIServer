package models

import (
	"github.com/astaxie/beego/orm"

	"github.com/hwiewie/APIServer/util/logs"
)

const (
	PublishTypeDeployment PublishType = iota
	PublishTypeService
	PublishTypeConfigMap
	PublishTypeSecret
	PublishTypePersistentVolumeClaim
	PublishTypeCronJob
	PublishTypeStatefulSet
	PublishTypeDaemonSet
	PublishTypeIngress
	PublishTypeHPA

	TableNamePublishStatus = "publish_status"
)

type publishStatusModel struct{}

type PublishType int32

// 記錄已發布模版信息
type PublishStatus struct {
	Id         int64       `orm:"auto" json:"id,omitempty"`
	Type       PublishType `orm:"index;type(integer)" json:"type,omitempty"`
	ResourceId int64       `orm:"index;column(resource_id)" json:"resourceId,omitempty"`
	TemplateId int64       `orm:"index;column(template_id);" json:"templateId,omitempty"`
	Cluster    string      `orm:"size(128);column(cluster)" json:"cluster,omitempty"`
}

func (*PublishStatus) TableName() string {
	return TableNamePublishStatus
}

func (*publishStatusModel) GetAll(publishType PublishType, resourceId int64) (publishStatus []PublishStatus, err error) {
	_, err = Ormer().
		QueryTable(new(PublishStatus)).
		Filter("ResourceId", resourceId).
		Filter("Type", publishType).
		All(&publishStatus)
	return
}

func (*publishStatusModel) GetByCluster(publishType PublishType, resourceId int64, cluster string) (publishStatus PublishStatus, err error) {
	err = Ormer().
		QueryTable(new(PublishStatus)).
		Filter("ResourceId", resourceId).
		Filter("Type", publishType).
		Filter("Cluster", cluster).
		One(&publishStatus)
	return
}

func (*publishStatusModel) Publish(m *PublishStatus) error {
	o := orm.NewOrm()
	qs := o.QueryTable(new(PublishStatus))
	err := o.Begin()
	if err != nil {
		logs.Error("(%v) begin transaction error.%v", m, err)
		return err
	}
	publishStatus := []PublishStatus{}
	count, err := qs.Filter("ResourceId", m.ResourceId).
		Filter("Type", m.Type).
		All(&publishStatus)
	if err != nil {
		return err
	}
	// 該資源未發布過
	if count == 0 {
		_, err := o.Insert(m)
		transactionError := o.Commit()
		if transactionError != nil {
			logs.Error("(%v) commit transaction error.%v", m, err)
		}
		return err
	}

	for _, state := range publishStatus {
		if state.Cluster == m.Cluster {
			// 模版已經發布過，不做任何操作
			if state.TemplateId == m.TemplateId {
				return nil
			} else { // 改集群已經被其他模版發布過，需要先刪除原來記錄
				_, err := o.Delete(&state)
				if err != nil {
					return err
				}

				_, err = o.Insert(m)
				if err != nil {
					transactionError := o.Rollback()
					if transactionError != nil {
						logs.Error("(%v) rollback transaction error.%v", m, err)
					}
					return err
				}
				transactionError := o.Commit()
				if transactionError != nil {
					logs.Error("(%v) commit transaction error.%v", m, err)
				}
				return err
			}
		}
	}
	// 未找到已發布的機房，可以直接發布
	_, err = o.Insert(m)
	transactionError := o.Commit()
	if transactionError != nil {
		logs.Error("(%v) commit transaction error.%v", m, err)
	}
	return err
}

func (*publishStatusModel) DeleteById(id int64) (err error) {
	v := PublishStatus{Id: id}
	// ascertain id exists in the database
	if err = Ormer().Read(&v); err == nil {
		_, err = Ormer().Delete(&v)
		return err
	}
	return
}

func (p *publishStatusModel) Add(id int64, tplId int64, cluster string, publishType PublishType) error {
	// 添加發布狀態
	publishStatus := PublishStatus{
		ResourceId: id,
		TemplateId: tplId,
		Type:       publishType,
		Cluster:    cluster,
	}
	err := p.Publish(&publishStatus)
	if err != nil {
		logs.Error("publish publishStatus (%v) to db error.%v", publishStatus, err)
		return err
	}
	return nil
}
