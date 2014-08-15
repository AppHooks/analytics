package models

import (
	"github.com/jinzhu/gorm"
)

type Model interface {
	ToMap() map[string]interface{}
}

type BaseModel struct {
	db    *gorm.DB
	model Model
}

func (m *BaseModel) Save() {
	var db = m.db
	if db.NewRecord(m.model) {
		db.Create(m.model)
	} else {
		db.Save(m.model)
	}
}

func NewBaseModel(db *gorm.DB, model Model) BaseModel {
	return BaseModel{db, model}
}
