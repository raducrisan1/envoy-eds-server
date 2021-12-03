package main

import (
	"context"
	"errors"
	"sync"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

type EdsTarget struct {
	Address string `json:"address"`
	Port    uint32 `json:"port"`
}

type KeyedEdsTarget struct {
	Key     string
	Address string `json:"address"`
	Port    uint32 `json:"port"`
}

type HttpServer struct {
	AllTargets map[string]EdsTarget
	mutex      *sync.Mutex
	DataCache  *cache.SnapshotCache
	Eds        *EdsResource
}

func NewHttpServer() *HttpServer {
	result := HttpServer{}
	result.Initialize()
	return &result
}

func (s *HttpServer) Initialize() {
	s.AllTargets = make(map[string]EdsTarget)
	s.mutex = &sync.Mutex{}
}

func (s *HttpServer) List() []EdsTarget {
	rsp := make([]EdsTarget, len(s.AllTargets))
	i := 0
	for _, val := range s.AllTargets {
		rsp[i] = val
		i++
	}
	return rsp
}

func (s *HttpServer) Get(key string) (*EdsTarget, bool) {
	if t, ok := s.AllTargets[key]; ok {
		return &t, true
	}
	return nil, false
}

func toEdsTarget(src KeyedEdsTarget) EdsTarget {
	return EdsTarget{
		Address: src.Address,
		Port:    src.Port,
	}
}

func (s *HttpServer) Post(key string, target KeyedEdsTarget) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.AllTargets[key] = toEdsTarget(target)
	snapshot := s.Eds.GenerateSnapshot()

	return (*s.DataCache).SetSnapshot(context.Background(), s.Eds.NodeId, snapshot)
}

func (s *HttpServer) Delete(key string) error {
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

func (s *HttpServer) GetEndpoints() []*endpoint.LbEndpoint {
	result := make([]*endpoint.LbEndpoint, len(s.AllTargets))
	i := 0
	for _, val := range s.AllTargets {
		ep := endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  val.Address,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: val.Port,
								},
							},
						},
					},
				},
			},
		}
		result[i] = &ep
		i++
	}
	return result
}
