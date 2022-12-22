package axtools

func IfThenElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
}

func IfThen(condition bool, a interface{}) interface{} {
	if condition {
		return a
	}
	return nil
}
