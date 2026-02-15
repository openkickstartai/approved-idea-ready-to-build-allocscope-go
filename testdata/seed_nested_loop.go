package testdata

func nestedLoop() {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			buf := make([]byte, i*j)
			_ = buf
		}
	}
}
