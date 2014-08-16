package models

import (
	"encoding/json"
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

func (s *Service) GetConfiguration() map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal([]byte(s.Configuration), &result)
	return result
}

func NewService(db *gorm.DB, name string, configuration map[string]interface{}) *Service {
	marshalConfiguration, _ := json.Marshal(configuration)
	service := Service{
		Name:          name,
		Configuration: string(marshalConfiguration),
	}
	service.BaseModel = NewBaseModel(db, &service)
	return &service
}
