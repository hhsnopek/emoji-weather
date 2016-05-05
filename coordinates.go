package emoji_weather

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

type coordinates struct {
	lng string
	lat string
}

type data struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

func getCoordinates(city string) (*coordinates, error) {
	const URL = "http://maps.googleapis.com/maps/api/geocode/json?address="
	var data = data{}
	var c = &coordinates{}

	resp, err := http.Get(URL + url.QueryEscape(city))
	if err != nil {
		return &coordinates{}, err
	}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return &coordinates{}, err
	}

	if err = resp.Body.Close(); err != nil {
		return &coordinates{}, err
	}
	defer resp.Body.Close()

	if len(data.Results) > 0 && data.Status == "OK" {
		loc := data.Results[0].Geometry.Location
		c.lng = strconv.FormatFloat(loc.Lng, 'f', -1, 64)
		c.lat = strconv.FormatFloat(loc.Lat, 'f', -1, 64)
	} else {
		return &coordinates{}, errors.New("It appears your city does not exist! 👽")
	}

	return c, nil
}
