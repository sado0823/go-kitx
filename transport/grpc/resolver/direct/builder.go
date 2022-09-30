package direct

import (
	"strings"

	googleResolver "google.golang.org/grpc/resolver"
)

const (
	Scheme = "direct"
)

type builder struct{}

//	NewBuilder example direct://<authority>/127.0.0.1:9000,127.0.0.2:9000
func NewBuilder() googleResolver.Builder {
	return &builder{}
}

func (b *builder) Build(target googleResolver.Target, cc googleResolver.ClientConn, opts googleResolver.BuildOptions) (googleResolver.Resolver, error) {
	addrs := make([]googleResolver.Address, 0)
	paths := strings.Split(strings.TrimPrefix(target.URL.Path, "/"), ",")

	for _, path := range paths {
		addrs = append(addrs, googleResolver.Address{
			Addr: path,
		})
	}

	err := cc.UpdateState(googleResolver.State{
		Addresses: addrs,
	})
	if err != nil {
		return nil, err
	}

	return newResolver(), nil
}

func (b *builder) Scheme() string {
	return Scheme
}
