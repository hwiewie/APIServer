package models_test

import (
	"encoding/json"
	"testing"

	"github.com/hwiewie/APIServer/common"
	"github.com/hwiewie/APIServer/models"
)

func TestListApp(t *testing.T) {
	apps, err := models.AppModel.List(&common.QueryParam{
		PageNo:   1,
		PageSize: 10,
		Query: map[string]interface{}{
			"namespace_id": 2,
			"deleted":      false,
		},
	}, true, 1)
	if err != nil {
		t.Error(err)
	}
	app, err := json.Marshal(apps)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(app))
}

func TestCreatApp(t *testing.T) {
	id, err := models.AppModel.Add(&models.App{
		Name: "example",
		Namespace: &models.Namespace{
			Id: 1,
		},
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(string(id))
}
