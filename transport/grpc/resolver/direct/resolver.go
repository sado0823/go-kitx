package direct

import (
	googleResolver "google.golang.org/grpc/resolver"
)

type resolver struct {
}

func newResolver() googleResolver.Resolver {
	return &resolver{}
}

func (r *resolver) ResolveNow(googleResolver.ResolveNowOptions) {
	
}

func (r *resolver) Close() {

}
