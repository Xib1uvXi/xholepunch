package predictor

// LinearPortPredictor is a port predictor that predicts the next port by adding 1 to the current port.
type LinearPortPredictor struct {
	port int
}

func NewLinearPortPredictor(port int) *LinearPortPredictor {
	return &LinearPortPredictor{port: port}
}

func (d *LinearPortPredictor) NextPort() int {
	d.port++
	return d.port
}
