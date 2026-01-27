package densitysurvey

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"strconv"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/google"
)

const (
	MaxSamples = 256
)

type DensitySpreadsheet struct {
	spreadsheet *google.GSpreadsheet
}

func NewDensitySpreadsheet(sheetid string, ss *google.GSpreadsheetsService) (*DensitySpreadsheet, error) {
	var (
		s *google.GSpreadsheet
		err error
	)
	f := func() (*google.GSpreadsheet, error) {
		return ss.Sheet(sheetid)
	}

	s, err = google.RateLimit(f, 30 * time.Second)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("Unable to load sheet %s", sheetid))
	}

	return &DensitySpreadsheet{
		spreadsheet: s,
	}, nil
}

func (ds *DensitySpreadsheet) GetMeasurements() ([]Measurement, error) {
	var reterr error = nil
	ret := []Measurement{}

	for _, sheet := range ds.spreadsheet.GetSheets() {
		if m, err := ds.parseSheet(sheet.Properties.Title); err != nil {
			err = errors.Join(reterr, err)
		} else {
			ret = append(ret, m)
		}
	}

	return ret, reterr
}

func (ds *DensitySpreadsheet) parseSheet(name string) (Measurement, error) {
	m := Measurement{
		Name: name,
		DataPoints: make([]DataPoint, 10),
	}

	endcell := fmt.Sprintf("H%d", MaxSamples)

	// get the cmdrname and project
	data, err := ds.spreadsheet.ReadRange(name, "A1", endcell)
	if err != nil {
		return m, err
	}
	parts := strings.Split(data.Values[0][0].(string), " - ")
	if len(parts) == 2 {
		m.CMDR = parts[0]
		m.Project = parts[1]
	}

	defint := func(in any) int {
		if x, err := strconv.Atoi(in.(string)); err == nil {
			return x
		}
		fmt.Printf("Can't int: %v\n", in)
		return 0
	}
	deffloat := func(in any) float32 {
		if x, err := strconv.ParseFloat(in.(string), 32); err == nil {
			return float32(x)
		}
		fmt.Printf("Can't float: %v\n", in)
		return 0
	}

	for i, row := range data.Values {
		// headers and stuff
		if i <5 {
			continue
		}
		if len(row)<4 {
			break
		}
		dp := DataPoint{
			SystemName: row[0].(string),
			ZSample: defint(row[1]),
			Count: defint(row[2]),
			MaxDistance: deffloat(row[3]),
		}
		m.DataPoints = append(m.DataPoints, dp)
	}

	return m, nil
}
