package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8890/api/products", nil)
	req.Header.Add("X-Api-Key", `a6e1226d621b65992f20b82a835ff688843ed3c72f8d6f9c796e959cf4d59490`)
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
