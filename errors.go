package cmap

type IllegalParameterError struct {
	msg string
}

func (i IllegalParameterError) Error() string {
	return i.msg
}

func newIllegalParameterError(msg string) IllegalParameterError {
	return IllegalParameterError{
		msg: "cmap: illegal parameter: " + msg,
	}
}


