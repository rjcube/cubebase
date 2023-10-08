package pgx

type CubePgExecError struct {
	Msg string
}

func (e *CubePgExecError) Error() string {
	return e.Msg
}
