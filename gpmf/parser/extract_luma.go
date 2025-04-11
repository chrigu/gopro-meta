package parser

import (
	"gopro/internal"
)

type Luma struct {
	Luminance uint8
}

func ParseLumaData(klvs []KLV) [][]Luma {
	return extractSensorData(klvs,
		"Average luminance",
		extractLumaData)
}

func extractLumaData(klv KLV) []Luma {

	var payload [][]uint8

	for _, child := range klv.Children {
		// log("Processing child:", child.FourCC)

		switch child.FourCC {
		case "YAVG":
			internal.Log("YAVG found")
			payload = readPayload(child).([][]uint8)

		default:
			//log("Unknown FourCC", klv.FourCC)
		}
	}

	lumminanceData := make([]Luma, len(payload))
	for i := range payload {
		lumminanceData[i] = Luma{
			Luminance: payload[i][0],
		}
	}

	return lumminanceData
}
