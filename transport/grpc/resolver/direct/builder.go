package direct

import (
	"strings"

	"github.com/sado0823/go-kitx/kit/log"

	googleResolver "google.golang.org/grpc/resolver"
)

const (
	Scheme = "direct"
)

func init() {
	googleResolver.Register(NewBuilder())
}

type builder struct{}

//	NewBuilder example direct://<authority>/127.0.0.1:9000,127.0.0.2:9000
func NewBuilder() googleResolver.Builder {
	return &builder{}
}

func (b *builder) Build(target googleResolver.Target, cc googleResolver.ClientConn, opts googleResolver.BuildOptions) (googleResolver.Resolver, error) {
	addrs := make([]googleResolver.Address, 0)
	log.Infof("direct build:%s", target.URL.Path)
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
