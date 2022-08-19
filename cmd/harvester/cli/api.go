package cli

// func getJokeData(baseAPI string) []byte {
// 	request, err := http.NewRequest(
// 		http.MethodGet, //method
// 		baseAPI,        //url
// 		nil,            //body
// 	)

// 	if err != nil {
// 		log.Printf("Could not request a dadjoke. %v", err)
// 	}

// 	request.Header.Add("Accept", "application/json")
// 	request.Header.Add("User-Agent", "Dadjoke CLI (https://github.com/example/dadjoke)")

// 	response, err := http.DefaultClient.Do(request)
// 	if err != nil {
// 		log.Printf("Could not make a request. %v", err)
// 	}

// 	responseBytes, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		log.Printf("Could not read response body. %v", err)
// 	}

// 	return responseBytes
// }
