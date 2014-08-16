package models

import (
	"github.com/jinzhu/gorm"
)

type Service struct {
	BaseModel
	Id            int64
	Name          string
	Configuration string
}

func (s *Service) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name": s.Name,
	}
}

func NewService(db *gorm.DB, name string, configuration map[string]interface{}) *Service {
	service := Service{
		Name: name,
	}
	service.BaseModel = NewBaseModel(db, &service)
	return &service
}
