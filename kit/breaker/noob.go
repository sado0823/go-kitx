package breaker

type noob struct {
}

func (g *noob) MarkSuccess() {
}

func (g *noob) MarkFail() {

}

func (g *noob) Allow() error {
	return nil
}
