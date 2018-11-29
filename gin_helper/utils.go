package gin_helper

func Limit(l, max, _default int) int {
	if l <= 0 {
		return _default
	}

	if l > max {
		return max
	}

	return l
}
