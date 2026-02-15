package testdata

import "fmt"

func deferExample() {
	defer func() {
		fmt.Println("cleanup")
	}()
}

func goExample() {
	go func() {
		buf := make([]byte, 256)
		_ = buf
	}()
}
