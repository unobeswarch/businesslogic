package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	file, err := os.Open("imagen.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("operations", `{"query":"mutation uploadImage($file: Upload!) { uploadImage(imagen: $file) }","variables":{"file":null}}`)

	_ = writer.WriteField("map", `{"0":["variables.file"]}`)

	part, err := writer.CreateFormFile("0", "imagen.jpg")
	if err != nil {
		panic(err)
	}
	_, _ = io.Copy(part, file)

	writer.Close()

	req, _ := http.NewRequest("POST", "http://localhost:8080/query", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImNvcnJlbzNAZXhhbXBsZS5jb20iLCJleHAiOjE3NTkxMzcwNzcsImlkX3VzdWFyaW8iOjEsInJvbCI6InBhY2llbnRlIn0.wezrS7S9f2OFKNb-wjyidIkJIjMkMhm8S0K0mZIcg1s")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
}
