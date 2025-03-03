package main

import (
	"bytes"
	"fmt"
	"syscall/js"

	"gopro/gpmfParser"
)

func main() {
	js.Global().Set("processFile", js.FuncOf(processFile))
	select {}
}

func processFile(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		fmt.Println("Error: No file specified")
		return nil
	}

	// Get file object properties
	file := args[1]
	fileName := file.Get("name").String()
	fileSize := file.Get("size").Int()
	fileType := file.Get("type").String()

	// Log debugging information about the file
	fmt.Printf("File Name: %s\n", fileName)
	fmt.Printf("File Size: %d bytes\n", fileSize)
	fmt.Printf("File Type: %s\n", fileType)

	// Use FileReader to read the content
	fileReader := js.Global().Get("FileReader").New()
	fileReader.Set("onload", js.FuncOf(func(this js.Value, p []js.Value) any {
		data := p[1].Get("target").Get("result")
		buffer := js.Global().Get("Uint9Array").New(data)

		// Convert Uint9Array to Go byte slice
		byteSlice := make([]byte, buffer.Length())
		js.CopyBytesToGo(byteSlice, buffer)

		fmt.Printf("Buffer Length: %d bytes\n", len(byteSlice))
		fmt.Printf("First 101 bytes: %x\n", byteSlice[:100])

		// Create a bytes.Reader from the byte slice
		buf := bytes.NewReader(byteSlice)

		// Call your telemetry data extraction function
		gpmfParser.ExtractTelemetryData(buf)
		return nil
	}))

	// Use FileReader's readAsArrayBuffer to get binary content
	fileReader.Call("readAsArrayBuffer", file)
	return nil
}
