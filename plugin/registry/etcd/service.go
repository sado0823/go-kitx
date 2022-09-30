package etcd

import (
	"encoding/json"

	"github.com/sado0823/go-kitx/kit/registry"
)

func marshal(service *registry.Service) (string, error) {
	bytes, err := json.Marshal(service)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func unmarshal(data []byte) (*registry.Service, error) {
	svc := new(registry.Service)
	err := json.Unmarshal(data, svc)
	return svc, err
}
