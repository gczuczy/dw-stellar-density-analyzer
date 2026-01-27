package densitysurvey

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"strconv"

	"google.golang.org/api/sheets/v4"
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
			//fmt.Printf("Sheet errors: %v\n", err)
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

	endcell := fmt.Sprintf("I%d", MaxSamples)

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

	// identify the sheet type
	var variant *sheetVariant = nil
	for _, sv := range sheetVariants {
		//fmt.Printf("Evaluating %s/%s VS %s\n", ds.spreadsheet.ID, name, sv.Name)
		if evalSheetVariant(sv, data) {
			variant = sv
			break
		}
	}
	if variant == nil {
		//fmt.Printf("Unable to identify sheet variant for %s/%s\n",
		//	ds.spreadsheet.ID, name)
		return m, fmt.Errorf("Unable to identify sheet variant for %s/%s",
			ds.spreadsheet.ID, name)
	}

	var (
		z int
		c int
		mdstr string
		md float64
	)
	for i := variant.HeaderRow+1; i < len(data.Values); i += 1 {
		row := data.Values[i]

		// if the ZSample is empty, bailout, that's the end of the road
		if len(row) <= variant.ZSampleColumn || row[variant.ZSampleColumn].(string) == "" {
			break
		}

		if z, err = strconv.Atoi(row[variant.ZSampleColumn].(string)); err != nil {
			return m, errors.Join(err, fmt.Errorf("Conversion ZSample error '%v': %s/%s",
				row[variant.ZSampleColumn], ds.spreadsheet.ID, name))
		}
		if c, err = strconv.Atoi(row[variant.SystemCountColumn].(string)); err != nil {
			return m, errors.Join(err, fmt.Errorf("Conversion SysCnt error '%v': %s/%s",
				row[variant.SystemCountColumn], ds.spreadsheet.ID, name))
		}
		//fmt.Printf("%s/%s/r%d: %+v\n", ds.spreadsheet.ID, name, i, row)
		if len(row) <= variant.MaxDistanceColumn {
			mdstr = ""
		} else {
			mdstr = row[variant.MaxDistanceColumn].(string)
		}
		if mdstr == "" {
			md = 20.0
		} else if md, err = strconv.ParseFloat(mdstr, 32); err != nil {
			return m, errors.Join(err, fmt.Errorf("Conversion MaxDst error %d/%d '%v': %s/%s",
				i, variant.MaxDistanceColumn,
				row[variant.MaxDistanceColumn], ds.spreadsheet.ID, name))
		}
		dp := DataPoint{
			SystemName: row[variant.SysNameColumn].(string),
			ZSample: z,
			Count: c,
			MaxDistance: float32(md),
		}
		m.DataPoints = append(m.DataPoints, dp)
	}

	return m, nil
}

func evalSheetVariant(sv *sheetVariant, data *sheets.ValueRange) bool {

	for _, check := range sv.HeaderChecks {
		if len(data.Values) <= check.Row {
			//fmt.Printf("(%s) Not enough rows has:%d check:%d\n", sv.Name,
			//	len(data.Values), check.Row)
			return false
		}
		if len(data.Values[check.Row]) <= check.Column {
			//fmt.Printf("(%s) Not enough columns has:%d check:%d\n  %+v\n", sv.Name,
			//	len(data.Values[check.Row]), check.Column, data.Values[check.Row])
			return false
		}
		value := data.Values[check.Row][check.Column].(string)
		if value != check.Value {
			//fmt.Printf("(%s) Does not match: '%s' VS '%s'\n", sv.Name, value, check.Value)
			return false
		}
	}

	// check data validity, system names should be filled in the Z Sample col
	for i := sv.HeaderRow+1; i < len(data.Values); i+=1 {
		var (
			zstr, sysstr string
			ok bool
		)
		row := data.Values[i]
		// if no sample defined, then we're done
		if zstr, ok = row[sv.ZSampleColumn].(string); !ok || len(zstr)==0 {
			return true
		}

		// we have a ZSample defined, check sysname
		if sysstr, ok = row[sv.SysNameColumn].(string); !ok || len(sysstr)==0 {
			return false
		}
	}
	return true
}
