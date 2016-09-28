package tsdb

import (
	"github.com/grafana/grafana/pkg/components/simplejson"
	"gopkg.in/guregu/null.v3"
)

type Query struct {
	RefId         string
	Model         *simplejson.Json
	Depends       []string
	DataSource    *DataSourceInfo
	Results       []*TimeSeries
	Exclude       bool
	MaxDataPoints int64
	IntervalMs    int64
}

type QuerySlice []*Query

type Request struct {
	TimeRange *TimeRange
	Queries   QuerySlice
}

type Response struct {
	BatchTimings []*BatchTiming          `json:"timings"`
	Results      map[string]*QueryResult `json:"results"`
}

type DataSourceInfo struct {
	Id                int64
	Name              string
	PluginId          string
	Url               string
	Password          string
	User              string
	Database          string
	BasicAuth         bool
	BasicAuthUser     string
	BasicAuthPassword string
}

type BatchTiming struct {
	TimeElapsed int64
}

type BatchResult struct {
	Error        error
	QueryResults map[string]*QueryResult
	Timings      *BatchTiming
}

type QueryResult struct {
	Error  error           `json:"error"`
	RefId  string          `json:"refId"`
	Series TimeSeriesSlice `json:"series"`
}

type TimeSeries struct {
	Name   string           `json:"name"`
	Points TimeSeriesPoints `json:"points"`
}

type TimePoint [2]null.Float
type TimeSeriesPoints []TimePoint
type TimeSeriesSlice []*TimeSeries

func NewTimePoint(value float64, timestamp float64) TimePoint {
	return TimePoint{null.FloatFrom(value), null.FloatFrom(timestamp)}
}

func NewTimeSeriesPointsFromArgs(values ...float64) TimeSeriesPoints {
	points := make(TimeSeriesPoints, 0)

	for i := 0; i < len(values); i += 2 {
		points = append(points, NewTimePoint(values[i], values[i+1]))
	}

	return points
}

func NewTimeSeries(name string, points TimeSeriesPoints) *TimeSeries {
	return &TimeSeries{
		Name:   name,
		Points: points,
	}
}
