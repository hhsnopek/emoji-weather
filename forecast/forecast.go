package forecast

import (
	"github.com/itsabot/abot/shared/datatypes"
	forecast "github.com/mlbright/forecast/v2"
)

type Forecast struct {
	API_KEY string
	City    *dt.City
	Lng     string
	Lat     string
	Time    string
}

type TwentyFourHours struct {
	now       *forecast.Forecast
	morning   *forecast.Forecast
	afternoon *forecast.Forecast
	evening   *forecast.Forecast
}

func (f *Forecast) Now() (*forecast.Forecast, error) {
	now, err := f.At(f.Time)
	if err != nil {
		return &forecast.Forecast{}, err
	}

	return now, err
}

func (f *Forecast) At(time string) (*forecast.Forecast, error) {
	st, err := forecast.Get(f.API_KEY, f.Lng, f.Lat, time, forecast.US)
	if err != nil {
		return nil, err
	}

	return st, nil
}

//TODO
// calculate times for morning, afternoon & evening
func (f *Forecast) Get24Hours() (*TwentyFourHours, error) {
	now, err := f.Now()
	if err != nil {
		return &TwentyFourHours{}, err
	}

	morning, err := f.At("")
	if err != nil {
		return &TwentyFourHours{}, err
	}

	afternoon, err := f.At("")
	if err != nil {
		return &TwentyFourHours{}, err
	}

	evening, err := f.At("")
	if err != nil {
		return &TwentyFourHours{}, err
	}

	tfh := &TwentyFourHours{
		now:       now,
		morning:   morning,
		afternoon: afternoon,
		evening:   evening,
	}

	return tfh, nil
}
