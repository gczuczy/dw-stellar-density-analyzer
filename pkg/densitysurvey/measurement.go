package densitysurvey

import (
	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/edsm"
)

var (
	// in cli-mode having a global state is a workaround
	edsms *edsm.EDSM
)

type Measurement struct {
	CMDR string
	Project string
	Name string
	DataPoints []DataPoint
}

type DataPoint struct {
	X float32
	Y float32
	Z float32
	SystemName string
	ZSample int
	Count int
	MaxDistance float32
}

func (m *Measurement) LookupNames() error {

	if edsms == nil {
		edsms = edsm.New()
	}

	names := make([]string, 0, len(m.DataPoints))
	for _, dp := range m.DataPoints {
		names = append(names, dp.SystemName)
	}

	lookupres, err := edsms.Systems(names)
	if err != nil {
		return err
	}

	// and correlate names
	for i, dp := range m.DataPoints {
		for _, sys := range lookupres {
			if sys.Name == dp.SystemName {
				m.DataPoints[i].X = sys.Coords.X
				m.DataPoints[i].Y = sys.Coords.Z
				m.DataPoints[i].Z = sys.Coords.Y
				break
			}
		}
	}

	return nil
}
