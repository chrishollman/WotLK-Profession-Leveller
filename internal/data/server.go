package data

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"go.uber.org/zap"
)

type Server struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	AHIds  []int  `json:"auctionhouseids"`
}

type ServerStore struct {
	dataslice []Server
	logger    *zap.SugaredLogger
}

func NewServerStore(logger *zap.SugaredLogger) *ServerStore {
	var data []Server

	bytes, err := os.ReadFile("data/servers.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	return &ServerStore{
		dataslice: data,
		logger:    logger,
	}
}

func (s *ServerStore) GetAll() []string {
	res := make([]string, len(s.dataslice))
	for _, v := range s.dataslice {
		res = append(res, v.Name)
	}
	return res
}

func (s *ServerStore) GetByName(name string) (*Server, error) {
	for _, v := range s.dataslice {
		if strings.EqualFold(name, v.Name) {
			return &v, nil
		}
	}

	return nil, errors.New("unable to locate server with that name")
}

func (s *ServerStore) Validate(server, region string) bool {
	for _, v := range s.dataslice {
		if strings.EqualFold(v.Name, server) && strings.EqualFold(v.Region, region) {
			return true
		}
	}

	return false
}
