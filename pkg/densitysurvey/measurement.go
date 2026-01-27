package densitysurvey

import (
)

type Measurement struct {
	CMDR string
	Project string
	Name string
	DataPoints []DataPoint
}

type DataPoint struct {
	X int
	Y int
	Z int
	SystemName string
	ZSample int
	Count int
	MaxDistance float32
}
