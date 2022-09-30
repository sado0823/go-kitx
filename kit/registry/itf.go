package registry

import "context"

type (
	Service struct {
		// ID unique id
		ID string `json:"id"`
		// Name can be same with the same app
		Name     string            `json:"name"`
		Version  string            `json:"version"`
		Metadata map[string]string `json:"metadata"`
		// Endpoints are endpoint addresses of the service instance.
		// schema:
		//   http://127.0.0.1:8000?isSecure=false
		//   grpc://127.0.0.1:9000?isSecure=false
		Endpoints []string `json:"endpoints"`
	}

	Registrar interface {
		Register(ctx context.Context, svc *Service) error
		Deregister(ctx context.Context, svc *Service) error
	}

	Watcher interface {
		// Next return all services if services not empty at the first time, OR block until service list change OR ctx timeout
		Next() ([]*Service, error)
		Stop() error
	}

	Discovery interface {
		Get(ctx context.Context, name string) ([]*Service, error)
		Watch(ctx context.Context, name string) (Watcher, error)
	}
)
