package testdata

func loopAlloc() {
	for i := 0; i < 100; i++ {
		s := make([]byte, 1024)
		_ = s
	}
}
