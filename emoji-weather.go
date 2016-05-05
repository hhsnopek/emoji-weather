package emoji_weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	forecast "github.com/hhsnopek/emoji-weather/forecast"
	"github.com/itsabot/abot/shared/datatypes"
	"github.com/itsabot/abot/shared/language"
	"github.com/itsabot/abot/shared/nlp"
	"github.com/itsabot/abot/shared/plugin"
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

var p *dt.Plugin

func init() {
	trigger := &nlp.StructuredInput{
		Commands: []string{"what", "show", "tell", "is", "will", "do"},
		Objects: []string{"weather", "temperature", "temp", "outside",
			"rain", "tomorrow"},
	}

	fns := &dt.PluginFns{Run: Run, FollowUp: FollowUp}
	p, err := plugin.New(REPO, trigger, fns)
	if err != nil {
		log.Fatal(err)
	}

	p.Vocab = dt.NewVocab(
		dt.VocabHandler{
			Fn: getWeather,
			Trigger: &nlp.StructuredInput{
				Commands: []string{"what", "show", "tell"},
				Objects:  []string{"weather"},
			},
		},
		dt.VocabHandler{
			Fn: getTemp,
			Trigger: &nlp.StructuredInput{
				Commands: []string{"what", "show", "tell"},
				Objects:  []string{"temperature", "temp"},
			},
		},
		dt.VocabHandler{
			Fn: getRain,
			Trigger: &nlp.StructuredInput{
				Commands: []string{"will", "do"},
				Objects:  []string{"rain", "umbrella"},
			},
		},
		dt.VocabHandler{
			Fn: getSpecificPoint,
			Trigger: &nlp.StructuredInput{
				Commands: []string{"what", "is"},
				Objects: []string{"weather", "temperature", "tomorrow",
					"sunday", "monday", "tuesday", "wednesday", "thursday",
					"friday", "saturday", "days", "next", "week"},
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

func getWeather(in *dt.Msg) (resp string) {
	city, err := getCity(in)
	if err != nil {
		return er(err)
	}

	coordinates, err := getCoordinates(city.Name)
	if err != nil {
		return er(err)
	}

	now := parseTime(in)

	weather := &forecast.Forecast{
		FORECAST_API_KEY,
		city,
		coordinates.lng,
		coordinates.lat,
		now,
	}

	f, err := weather.Now()
	if err != nil {
		return er(err)
	}

	maxTempTime := strings.Split(time.Unix(int(f.Currently.TemperatureMaxTime)),
		" ")[1]
	minTempTime := strings.Split(time.Unix(int(f.Currently.TemperatureMinTime)),
		" ")[1]

	current := fmt.Sprintf("It's currently %s and %.f° in %s with a high of %.f at %s and low of %.f at %s UTC.",
		f.Currently.Temperature,
		city,
		emojis[f.Currently.Icon],
		f.Currently.TemperatureMax,
		maxTempTime,
		f.Currently.TemperatureMin,
		minTempTime)

	var feelsLike, chanceOfRain, windSpeed, humidity, afternoon, evening string
	if f.Currently.Temperature != f.Currently.ApparentTemperature {
		feelsLike = fmt.Sprintf(" Yet it feels like %.f°",
			f.Currently.ApparentTemperature)
	}

	if f.Currently.Humidity != 0 {
		humidity = fmt.Sprintf(" The humidity is %.f.",
			f.Currently.Humidity)
	}

	if f.Currently.PrecipProbability != 0 {
		chanceOfRain = fmt.Sprintf(" There's a %.f chance of rain.",
			f.Currently.PrecipProbability)
	}

	if f.Currently.WindSpeed != 0 {
		windSpeed = fmt.Sprintf(" The wind speed is %.f mph.",
			f.Currently.WindSpeed)
	}

	return fmt.Sprint(current, feelsLike, humidity, chanceOfRain, windSpeed)
}

func getTemp(in *dt.Msg) (resp string) {
	city, lng, lat, err := getCityData(in)
	if err != nil {
		return er(err)
	}

	time := getTime(in)
	f, err := getForecastData(lng, lat, time)
	if err != nil {
		return er(err)
	}

	feelsLike := ""
	if f.Currently.Temperature != f.Currently.ApparentTemperature {
		feelsLike = fmt.Sprintf(" Yet it feels like %.f°",
			f.Currently.ApparentTemperature)
	}

	return fmt.Sprintf("It's %.f° right now in %s.%s",
		f.Currently.Temperature,
		city.Name,
		feelsLike)
}

func getRain(in *dt.Msg) (resp string) {
	city, err := getCity(in)
	if err != nil {
		return er(err)
	}

	coordinates, err := getCoordinates(city.Name)
	if err != nil {
		return er(err)
	}
	Time := ""

	weather := &forecast.Forecast{
		FORECAST_API_KEY,
		city,
		coordinates.lng,
		coordinates.lat,
	}

	f, err := weather.Now()
	if err != nil {
		return er(err)
	}

	return "Not Implemented"
}

func getSpecificPoint(in *dt.Msg) (resp string) {
	city, lng, lat, err := getCityData(in)
	if err != nil {
		return er(err)
	}

	time := parseTime(in)
	f, err := getForecastData(lng, lat, time)
	if err != nil {
		return er(err)
	}

	return "NOT IMPLEMENTED"
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
		if err := json.Unmarshal(mem.Val, city); err != nil {
			return nil, errors.New("Unable to unmarshal memory into city")
		}
	}

	if city == nil {
		return nil, errors.New("no cities found")
	}

	return city, nil
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
