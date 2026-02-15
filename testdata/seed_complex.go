package testdata

import "fmt"

type Server struct {
	handlers []func()
}

func (s *Server) Register() {
	for i := 0; i < 10; i++ {
		idx := i
		s.handlers = append(s.handlers, func() {
			fmt.Printf("handler %d\n", idx)
		})
	}
}

func (s *Server) Process(items []string) []error {
	var errs []error
	for _, item := range items {
		if item == "" {
			errs = append(errs, fmt.Errorf("empty item"))
		}
	}
	return errs
}
