package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8890/api/products", nil)
	req.Header.Add("X-Api-Key", `306150366f91434b986089c0e306eb7fda501c3bc2c43101549b11bb828aa517`)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Err: %+v\n", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Resp.Status: %d\n", resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Err: %+v\n", err)
	}
	bodyString := string(bodyBytes)

	fmt.Printf("Resp.Body: %s\n", bodyString)
}
