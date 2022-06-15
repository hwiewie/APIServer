package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	kapi "k8s.io/api/core/v1"

	"github.com/hwiewie/APIServer/util/hack"
)

const (
	TableNameStatefulset = "statefulset"
)

type statefulsetModel struct{}

type StatefulsetMetaData struct {
	Replicas  map[string]int32  `json:"replicas"`
	Resources map[string]string `json:"resources,omitempty"`
	Affinity  *kapi.Affinity    `json:"affinity,omitempty"`
	// 是否允許用戶使用特權模式，默認不允許,key 為容器名稱
	Privileged map[string]*bool `json:"privileged"`
}

type Statefulset struct {
	Id   int64  `orm:"auto" json:"id,omitempty"`
	Name string `orm:"unique;index;size(128)" json:"name,omitempty"`
	/* 存儲元數據
	{
	  "replicas": {
	    "K8S": 1
	  },
	  "privileged":{"nginx",true},
	  "affinity": {
	    "podAntiAffinity": {
	      "requiredDuringSchedulingIgnoredDuringExecution": [
	        {
	          "labelSelector": {
	            "matchExpressions": [
	              {
	                "operator": "In",
	                "values": [
	                  "xxx"
	                ],
	                "key": "app"
	              }
	            ]
	          },
	          "topologyKey": "kubernetes.io/hostname"
	        }
	      ]
	    }
	  },
	  "resources":{
			"cpuRequestLimitPercent": "50%", // cpu request和limit百分比，默認50%
			"memoryRequestLimitPercent": "100%", // memory request和limit百分比，默認100%
			"cpuLimit":"12",  // cpu限制，默認12個核
			"memoryLimit":"64" // 內存限制，默認64G
			"replicaLimit":"32" // 份數限制，默認32份
	  }
	}
	*/
	MetaData    string              `orm:"type(text)" json:"metaData,omitempty"`
	MetaDataObj StatefulsetMetaData `orm:"-" json:"-"`
	App         *App                `orm:"index;rel(fk)" json:"app,omitempty"`
	Description string              `orm:"null;size(512)" json:"description,omitempty"`
	OrderId     int64               `orm:"index;default(0)" json:"order"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	AppId int64 `orm:"-" json:"appId,omitempty"`
}

func (*Statefulset) TableName() string {
	return TableNameStatefulset
}

func (*statefulsetModel) GetNames(filters map[string]interface{}) ([]Statefulset, error) {
	var statefulsets []Statefulset
	qs := Ormer().
		QueryTable(new(Statefulset))

	if len(filters) > 0 {
		for k, v := range filters {
			qs = qs.Filter(k, v)
		}
	}
	_, err := qs.All(&statefulsets, "Id", "Name")

	if err != nil {
		return nil, err
	}

	return statefulsets, nil
}

func (*statefulsetModel) UpdateOrders(statefulsets []*Statefulset) error {
	if len(statefulsets) < 1 {
		return errors.New("statefulsets' length should greater than 0. ")
	}
	batchUpateSql := fmt.Sprintf("UPDATE `%s` SET `order_id` = CASE ", TableNameStatefulset)
	ids := make([]string, 0)
	for _, statefulset := range statefulsets {
		ids = append(ids, strconv.Itoa(int(statefulset.Id)))
		batchUpateSql = fmt.Sprintf("%s WHEN `id` = %d THEN %d ", batchUpateSql, statefulset.Id, statefulset.OrderId)
	}
	batchUpateSql = fmt.Sprintf("%s END WHERE `id` IN (%s)", batchUpateSql, strings.Join(ids, ","))

	_, err := Ormer().Raw(batchUpateSql).Exec()
	return err
}

func (*statefulsetModel) Add(m *Statefulset) (id int64, err error) {
	m.App = &App{Id: m.AppId}
	id, err = Ormer().Insert(m)
	return
}

func (*statefulsetModel) UpdateById(m *Statefulset) (err error) {
	v := Statefulset{Id: m.Id}
	// ascertain id exists in the database
	if err = Ormer().Read(&v); err == nil {
		m.App = &App{Id: m.AppId}
		_, err = Ormer().Update(m)
		return err
	}
	return
}

func (*statefulsetModel) GetById(id int64) (v *Statefulset, err error) {
	v = &Statefulset{Id: id}

	if err = Ormer().Read(v); err == nil {
		v.AppId = v.App.Id
		return v, nil
	}
	return nil, err
}

func (*statefulsetModel) GetParseMetaDataById(id int64) (v *Statefulset, err error) {
	v = &Statefulset{Id: id}

	if err = Ormer().Read(v); err == nil {
		v.AppId = v.App.Id
		err = json.Unmarshal(hack.Slice(v.MetaData), &v.MetaDataObj)
		if err != nil {
			logs.Error("parse statefulset metaData error.", v.MetaData)
			return nil, err
		}
		return v, nil
	}
	return nil, err
}

func (*statefulsetModel) DeleteById(id int64, logical bool) (err error) {
	v := Statefulset{Id: id}
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
