package cubebase

type CubeError struct {
	Msg string
}

func (e *CubeError) Error() string {
	return e.Msg
}
