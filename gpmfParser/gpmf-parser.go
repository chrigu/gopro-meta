package gpmfParser

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/fatih/color"
)

type KLV struct {
	FourCC     string
	DataType   int
	DataSize   uint32
	Repeat     uint32
	Payload    []byte
	Children   []KLV
	ParsedData []any
}

type GPS9 struct {
	Latitude  float32
	Longitude float32
	Altitude  float32
}

func ParseGPMF(data []byte) []KLV {
	var offset uint32 = 0
	var klvs []KLV = make([]KLV, 0)

	for {
		newOffset := readKLV(data, offset, &klvs)

		if newOffset <= offset { // Stops infinite loop when offset is not advancing
			fmt.Println("Error: Offset did not advance, stopping.")
			break
		}

		fmt.Println("Offset advanced to:", newOffset)
		offset = newOffset

		if offset >= uint32(len(data)) { // Ensures we don't read beyond available data
			// if offset >= 1000 { // Ensures we don't read beyond available data
			// PrintTree(klvs, "")
			break
		}
	}

	return klvs
}

func readKLV(data []byte, offset uint32, klvs *[]KLV) uint32 {
	// Check if enough bytes exist before reading

	dataOffset := offset + 8

	if dataOffset > uint32(len(data)) {
		fmt.Println("Error: Not enough data for KLV header")
		return offset + 8
	}

	klv := KLV{
		FourCC:   string(data[offset : offset+4]),
		DataType: int(data[offset+4]),
		DataSize: uint32(data[offset+5]),
		Repeat:   uint32(data[offset+6])<<8 | uint32(data[offset+7]),
		Children: make([]KLV, 0),
	}

	// Ensure payload does not exceed data slice
	if dataOffset+klv.DataSize*klv.Repeat > uint32(len(data)) {
		// fmt.Println("Error: Payload exceeds available data")
		return dataOffset
	}

	klv.Payload = data[dataOffset : dataOffset+klv.DataSize*klv.Repeat]
	*klvs = append(*klvs, klv)

	totalSize := klv.DataSize * klv.Repeat
	padding := (4 - (totalSize % 4)) % 4

	//fmt.Println("FourCC:", klv.FourCC, "DataType:", klv.DataType, "DataSize:", klv.DataSize, "Repeat:", klv.Repeat, "Padding:", padding)

	// Process nested KLV structures

	switch klv.DataType {
	case 0:
		fmt.Println("Processing nested KLV entries")
		nestedOffset := uint32(0) + padding

		// Process multiple nested KLVs inside the payload
		for nestedOffset < uint32(len(klv.Payload)) {
			var nestedKLVs []KLV
			nestedOffset = readKLV(klv.Payload, nestedOffset, &nestedKLVs)

			if len(nestedKLVs) > 0 {
				(*klvs)[len(*klvs)-1].Children = append((*klvs)[len(*klvs)-1].Children, nestedKLVs...)
			}
		}

		if klv.FourCC == "STRM" {
			fmt.Println("Stream KLV found")
			index := slices.IndexFunc((*klvs)[len(*klvs)-1].Children, func(child KLV) bool {
				return strings.TrimSpace(string(child.Payload)) == "GPS (Lat., Long., Alt., 2D, 3D, days, secs, DOP, fix)"
			})
			if index != -1 {
				gpsData := extractGpsData((*klvs)[len(*klvs)-1])
				parsedData := make([]any, len(gpsData))
				for i, v := range gpsData {
					parsedData[i] = v
				}
				(*klvs)[len(*klvs)-1].ParsedData = parsedData
			}

		}

	case int('b'): // int8_t
		fmt.Println("Type: int8_t")
	case int('B'): // uint8_t
		fmt.Println("Type: uint8_t")
	case int('c'): // ASCII character string
		fmt.Println("Type: ASCII character string")
		// use repeat
		fmt.Println("Payload:", string(klv.Payload))
		(*klvs)[len(*klvs)-1].ParsedData = []any{string(klv.Payload)}
	case int('d'): // double
		fmt.Println("Type: double (64-bit float)")
	case int('f'): // float
		fmt.Println("Type: float (32-bit float)")
	case int('F'): // FourCC
		fmt.Println("Type: FourCC (32-bit character key)")
	case int('G'): // UUID
		fmt.Println("Type: UUID (128-bit identifier)")
	case int('j'): // int64_t
		fmt.Println("Type: int64_t (64-bit signed integer)")
	case int('J'): // uint64_t
		fmt.Println("Type: uint64_t (64-bit unsigned integer)")
	case int('l'): // int32_t
		fmt.Println("Type: int32_t (32-bit signed integer)")
		scal := make([]int32, klv.Repeat)
		for i := range klv.Repeat {
			scal[i] = int32(binary.BigEndian.Uint32(klv.Payload[i*4 : i*4+4]))
		}
		(*klvs)[len(*klvs)-1].ParsedData = []any{scal}
	case int('L'): // uint32_t
		fmt.Println("Type: uint32_t (32-bit unsigned integer)")

	case int('q'): // Q15.16
		fmt.Println("Type: Q15.16 (fixed-point 32-bit number)")
	case int('Q'): // Q31.32
		fmt.Println("Type: Q31.32 (fixed-point 64-bit number)")
	case int('s'): // int16_t
		fmt.Println("Type: int16_t (16-bit signed integer)")
	case int('S'): // uint16_t
		fmt.Println("Type: uint16_t (16-bit unsigned integer)")
	case int('U'): // UTC Date and Time string
		fmt.Println("Type: UTC Date and Time string")
	case int('?'): // Complex structure
		fmt.Println("Type: Complex structure")
	default:
		fmt.Println("Unknown data type")
	}

	return dataOffset + klv.DataSize*klv.Repeat + padding
}

// parseDynamicStructure dynamically parses a buffer based on the format string
func parseDynamicStructure(data []byte, format string) ([]interface{}, error) {
	fmt.Println("Parsing dynamic structure with format:", format)
	offset := 0
	totalSize := len(data)
	values := []interface{}{} // Slice to store parsed values

	for i, char := range format {
		switch char {
		case 'l': // 32-bit signed integer
			if offset+4 > totalSize {
				fmt.Printf("Error: Not enough data for int32 at position %d\n", i)
				return nil, fmt.Errorf("Not enough data for int32 at position %d", i)
			}
			value := int32(binary.BigEndian.Uint32(data[offset : offset+4]))
			fmt.Printf("l[%d]: %d (int32)\n", i, value)
			values = append(values, value)
			offset += 4

		case 'S': // 16-bit unsigned integer
			if offset+2 > totalSize {
				fmt.Printf("Error: Not enough data for uint16 at position %d\n", i)
				return nil, fmt.Errorf("Not enough data for uint16 at position %d", i)
			}
			value := binary.BigEndian.Uint16(data[offset : offset+2])
			fmt.Printf("S[%d]: %d (uint16)\n", i, value)
			values = append(values, value)
			offset += 2

		case 'f': // 32-bit float
			if offset+4 > totalSize {
				fmt.Printf("Error: Not enough data for float32 at position %d\n", i)
				return nil, fmt.Errorf("Not enough data for float32 at position %d", i)
			}
			value := math.Float32frombits(binary.BigEndian.Uint32(data[offset : offset+4]))
			fmt.Printf("f[%d]: %f (float32)\n", i, value)
			values = append(values, value)
			offset += 4

		default:
			fmt.Printf("Unknown format character: %c\n", char)
			return nil, fmt.Errorf("Unknown format character: %c", char)
		}
	}

	// Calculate padding
	padding := (4 - (offset % 4)) % 4
	if padding > 0 && offset+int(padding) <= totalSize {
		fmt.Printf("Padding bytes: %d\n", padding)
		offset += int(padding)
	}

	fmt.Printf("Total bytes processed: %d\n", offset)
	return values, nil
}

func extractGpsData(klv KLV) []GPS9 {
	fmt.Println("Processing STRM children", len(klv.Children))

	var format string = ""
	var payload []byte = make([]byte, 0)
	var scale []int32

	for _, child := range klv.Children {
		fmt.Println("Processing child:", child.FourCC)

		switch child.FourCC {
		case "GPS9":
			color.Green("GPS9 found")
			payload = child.Payload

		case "TYPE":
			fmt.Println("TYPE found", child.ParsedData)
			format = string(child.Payload)

		case "SCAL":
			fmt.Println("SCAL found", child.ParsedData)
			if len(child.ParsedData) > 0 {
				if parsedScale, ok := child.ParsedData[0].([]int32); ok {
					scale = parsedScale
				} else {
					fmt.Println("Error: ParsedData is not of type []int32")
				}
			}
		default:
			//fmt.Println("Unknown FourCC", klv.FourCC)
		}
	}

	gpsRawData, err := parseDynamicStructure(payload, format) // todo get from gopro
	if err != nil {
		fmt.Println("Error parsing dynamic structure:", err)
	}

	gpsData := []GPS9{
		{
			Latitude:  float32(gpsRawData[0].(int32)) / float32(scale[0]),
			Longitude: float32(gpsRawData[1].(int32)) / float32(scale[1]),
			Altitude:  float32(gpsRawData[2].(int32)) / float32(scale[2]),
		},
	}

	return gpsData

}

func parseScaling(data []byte) float64 {
	return 0.0
}
