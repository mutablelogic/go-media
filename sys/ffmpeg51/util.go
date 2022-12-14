package ffmpeg

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}
