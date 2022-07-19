// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml
//
// The code under api/petstore/ has been generated from that specification.
package server

import "fmt"

type Server struct {
	// config Config
}

func (s Server) Run() {
	fmt.Println("Hello, world!")
}
