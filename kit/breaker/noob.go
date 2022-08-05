package breaker

type noob struct {
}

func (g *noob) Name() string {
	return "noob-breaker"
}

func (g *noob) MarkSuccess() {
}

func (g *noob) MarkFail() {

}

func (g *noob) Allow() error {
	return nil
}
