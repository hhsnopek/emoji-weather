package emoji_weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/itsabot/abot/shared/datatypes"
	"github.com/itsabot/abot/shared/language"
	"github.com/itsabot/abot/shared/nlp"
	"github.com/itsabot/abot/shared/plugin"
	forecast "github.com/mlbright/forecast/v2"
)

const REPO string = "github.com/hhsnopek/emoji-weather"
const GOOGLE_API_KEY string = "AIzaSyDtprPpmM6wG2pHNWCmV-STEldxhTujKbQ"
const FORECAST_API_KEY string = "65fd8a1c94201e85f157014630f9668e"

var emojis = map[string]string{
	"clear-day":           "☀️",
	"clear-night":         "☀",
	"rain":                "🌧",
	"snow":                "🌨",
	"sleet":               "🌨",
	"wind":                "💨",
	"fog":                 "🌫",
	"cloudy":              "☁",
	"partly-cloudy-day":   "⛅",
	"partly-cloudy-night": "⛅",
}

type weatherJSON struct {
	Description []string
	Temp        float64
	Humidity    int
}

type cityCoordinates struct {
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

var p *dt.Plugin

func init() {
	rand.Seed(time.Now().UnixNano())
	trigger := &nlp.StructuredInput{
		Commands: []string{"what", "show", "tell", "is"},
		Objects:  []string{"weather", "temperature", "temp", "outside"},
	}
	fns := &dt.PluginFns{Run: Run, FollowUp: FollowUp}
	var err error
	p, err = plugin.New(REPO, trigger, fns)
	if err != nil {
		log.Fatal(err)
	}
	p.Vocab = dt.NewVocab(
		dt.VocabHandler{
			Fn: getTemp,
			Trigger: &nlp.StructuredInput{
				Commands: []string{"what", "show", "tell"},
				Objects: []string{"weather", "temperature",
					"temp", "outside"},
			},
		},
	)
}

func Run(in *dt.Msg) (string, error) {
	return FollowUp(in)
}

func FollowUp(in *dt.Msg) (string, error) {
	return p.Vocab.HandleKeywords(in), nil
}

func getTemp(in *dt.Msg) (resp string) {
	city, err := getCity(in)
	if err != nil {
		return er(err)
	}
	lng, lat, err := getCoordinates(city)
	if err != nil {
		return er(err)
	}

	return getWeather(city, lng, lat)
}

func getCity(in *dt.Msg) (*dt.City, error) {
	cities, err := language.ExtractCities(p.DB, in)
	if err != nil {
		return nil, errors.New(
			"Unable to locate the city in our DB.")
	}
	city := &dt.City{}
	sm := buildStateMachine(in)
	if len(cities) >= 1 {
		city = &cities[0]
	} else if sm.HasMemory(in, "city") {
		mem := sm.GetMemory(in, "city")
		p.Log.Debug(mem)
		if err := json.Unmarshal(mem.Val, city); err != nil {
			p.Log.Debug("couldn't unmarshal mem into city", err)
			return nil, errors.New("Neo is in your computer and we aren't able to decode our memory into a city. 🕴")
		}
	}
	if city == nil {
		return nil, errors.New("no cities found")
	}
	return city, nil
}

func getCoordinates(city *dt.City) (string, string, error) {
	var lng, lat string
	const URL = "http://maps.googleapis.com/maps/api/geocode/json?address="
	var cc = cityCoordinates{}

	resp, err := http.Get(URL + url.QueryEscape(city.Name))
	if err != nil {
		return "", "", err
	}

	if err = json.NewDecoder(resp.Body).Decode(&cc); err != nil {
		return "", "", err
	}

	if err = resp.Body.Close(); err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if len(cc.Results) > 0 && cc.Status == "OK" {
		loc := cc.Results[0].Geometry.Location
		lng = strconv.FormatFloat(loc.Lng, 'f', -1, 64)
		lat = strconv.FormatFloat(loc.Lat, 'f', -1, 64)
	} else {
		return "", "", errors.New("It appears your city does not exist! 👽")
	}

	return lat, lng, nil

}

func getWeather(city *dt.City, lng, lat string) string {
	p.Log.Debug("getting weather for city", city.Name)

	f, err := forecast.Get(FORECAST_API_KEY, lng, lat, "now", forecast.US)
	if err != nil {
		return er(err)
	}

	emoji := emojis[f.Currently.Icon]
	feelsLike := ""

	if f.Currently.Temperature != f.Currently.ApparentTemperature {
		feelsLike = fmt.Sprintf(" Yet it feels like %.f°",
			f.Currently.ApparentTemperature)
	}

	ret := fmt.Sprintf("It's %.f° & %s right now in %s.%s",
		f.Currently.Temperature,
		emoji,
		city.Name,
		feelsLike)
	return ret
}

func buildStateMachine(in *dt.Msg) *dt.StateMachine {
	sm := dt.NewStateMachine(p)
	sm.SetStates([]dt.State{})
	sm.LoadState(in)
	return sm
}

func er(err error) string {
	return "🤔 Uh oh... " + err.Error()
}
