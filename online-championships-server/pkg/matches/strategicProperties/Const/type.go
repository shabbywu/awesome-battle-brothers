package Const

type Ambience struct {
	File         string  `json:"File"`
	Pan          bool    `json:"Pan"`
	Pitch        float64 `json:"Pitch"`
	RandomVolume bool    `json:"RandomVolume"`
	Volume       float64 `json:"Volume"`
}
