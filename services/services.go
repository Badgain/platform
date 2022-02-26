package services

import (
	"errors"

	"github.com/Badgain/platform/database/models"
)

var (
	ErrNoNeedColumn = errors.New("no need column in map")
)

type LinkedService struct {
	Title string
	Token string
}

type ServicesList []*LinkedService

func (s *ServicesList) Load(dbc models.DbConnection) error {
	exp := `
		select s.title, at.token from services s 
		join access_tokens at on at.service_id = s.id 
		where at.is_valid = true
	`
	result, err := dbc.Query(exp)
	if err != nil {
		return err
	}
	for _, row := range result {
		serv, err := QueryRowToService(row)
		if err != nil {
			return err
		}
		*s = append(*s, serv)
	}
	return nil
}

func (s *ServicesList) CheckService(name string, token string) bool {
	for _, v := range *s {
		if v.Title == name && v.Token == token {
			return true
		}
	}
	return false
}

func (s *ServicesList) GetByToken(token string) *LinkedService {
	for _, unit := range *s {
		if unit.Token == token {
			return unit
		}
	}
	return nil
}

func QueryRowToService(r models.QueryRow) (*LinkedService, error) {
	if r["title"] == nil || r["token"] == nil {
		return nil, ErrNoNeedColumn
	}
	s := LinkedService{
		Title: r["title"].(string),
		Token: r["token"].(string),
	}
	return &s, nil
}

func LoadServices(dbc models.DbConnection) (*ServicesList, error) {
	sl := &ServicesList{}
	err := sl.Load(dbc)
	if err != nil {
		return nil, err
	}
	return sl, nil
}
