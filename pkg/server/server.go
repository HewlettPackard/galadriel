package server

import "fmt"

type Server struct {
	//config Config
}

func (s Server) Run() {
	fmt.Println("Hello, world!")
}
