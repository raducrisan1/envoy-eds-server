package main

import (
	"errors"
	"sync"
)

type EdsTarget struct {
	Address string	`json:"address"`
	Port int		`json:"port"`
}

type KeyedEdsTarget struct {
	Key     string
	Address string	`json:"address"`
	Port 	int		`json:"port"`
}

type Server struct {
	AllTargets map[string]EdsTarget
	mutex      *sync.Mutex
}

func NewServer() *Server {
	result := Server{}
	result.Initialize()
	return &result
}

func (s *Server) Initialize() {
	s.AllTargets = make(map[string]EdsTarget)
	s.mutex = &sync.Mutex{}
}

func (s *Server) List() []EdsTarget {
	rsp := make([]EdsTarget, len(s.AllTargets))
	i := 0
	for _, val := range s.AllTargets {
		rsp[i] = val
		i++
	}
	return rsp
}

func (s *Server) Get(key string) (*EdsTarget, bool) {
	if t, ok := s.AllTargets[key]; ok {
		return &t, true
	}
	return nil, false
}

func toEdsTarget(src KeyedEdsTarget) EdsTarget {
	return EdsTarget{
		Address: src.Address,
		Port: src.Port,
	}
}

func (s *Server) Post(key string, target KeyedEdsTarget) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.AllTargets[key] = toEdsTarget(target)
	return nil
}

func (s *Server) Delete(key string) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.AllTargets[key]; ok {
		delete(s.AllTargets, key)
		return nil
	}
	return errors.New("the key does not exist")
}
