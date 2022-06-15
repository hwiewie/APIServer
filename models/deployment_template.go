package models

import (
	"time"
)

const (
	TableNameDeploymentTemplate = "deployment_template"
)

type deploymentTplModel struct{}

type DeploymentTemplate struct {
	Id          int64       `orm:"auto" json:"id,omitempty"`
	Name        string      `orm:"size(128)" json:"name,omitempty"`
	Template    string      `orm:"type(text)" json:"template,omitempty"`
	Deployment  *Deployment `orm:"index;rel(fk);column(deployment_id)" json:"deployment,omitempty"`
	Description string      `orm:"size(512)" json:"description,omitempty"`

	// TODO
	// 如果使用指針類型auto_now_add和auto_now可以自動生效,但是orm QueryRows無法對指針類型的time正常賦值，
	// 不使用指針類型創建時需要手動把創建時間設置為當前時間,更新時也需要處理創建時間
	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	DeploymentId int64            `orm:"-" json:"deploymentId,omitempty"`
	Status       []*PublishStatus `orm:"-" json:"status,omitempty"`
}

func (*DeploymentTemplate) TableName() string {
	return TableNameDeploymentTemplate
}

func (*deploymentTplModel) Add(m *DeploymentTemplate) (id int64, err error) {
	m.Deployment = &Deployment{Id: m.DeploymentId}
	now := time.Now()
	m.CreateTime = now
	m.UpdateTime = now
	id, err = Ormer().Insert(m)
	return
}

func (*deploymentTplModel) UpdateById(m *DeploymentTemplate) (err error) {
	v := DeploymentTemplate{Id: m.Id}
	// ascertain id exists in the database
	if err = Ormer().Read(&v); err == nil {
		_, err = Ormer().Update(m)
		return err
	}
	return
}

func (*deploymentTplModel) GetById(id int64) (v *DeploymentTemplate, err error) {
	v = &DeploymentTemplate{Id: id}

	if err = Ormer().Read(v); err == nil {
		_, err = Ormer().LoadRelated(v, "Deployment")
		if err == nil {
			v.DeploymentId = v.Deployment.Id
			return v, nil
		}
	}
	return nil, err
}

func (*deploymentTplModel) DeleteById(id int64, logical bool) (err error) {
	v := DeploymentTemplate{Id: id}
	// ascertain id exists in the database
	if err = Ormer().Read(&v); err == nil {
		if logical {
			v.Deleted = true
			_, err = Ormer().Update(&v)
			return err
		}
		_, err = Ormer().Delete(&v)
		return err
	}
	return
}

func (*deploymentTplModel) GetLatestDeptplByName(ns, app, deployment string) (v *DeploymentTemplate, err error) {
	v = &DeploymentTemplate{}
	// use orm
	qs := Ormer().QueryTable(new(DeploymentTemplate))
	err = qs.Filter("Deployment__App__Namespace__Name", ns).Filter("Deployment__App__Name", app).Filter("Name", deployment).Filter("Deleted", 0).OrderBy("-id").One(v)
	if err == nil {
		v.DeploymentId = v.Deployment.Id
		return v, nil
	}
	return nil, err
}
