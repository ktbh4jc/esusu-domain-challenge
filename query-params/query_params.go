package query_params

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Lon   float64 `json:"lon"`
	Lat   float64 `json:"lat"`
	Query string  `json:"query"`
}

func ExtractParams(c *gin.Context) (*QueryParams, error) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		if c.Query("lat") == "" {
			lat = 0
		} else {
			return nil, err
		}
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		if c.Query("lon") == "" {
			lon = 0
		} else {
			return nil, err
		}
	}
	queryParams := &QueryParams{
		Lat:   lat,
		Lon:   lon,
		Query: c.Query("query"),
	}
	return queryParams, nil
}
