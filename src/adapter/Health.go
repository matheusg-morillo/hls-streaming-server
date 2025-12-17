package adapter

import (
	"matheusflix/hls-streaming-server/src/domain"
	"matheusflix/hls-streaming-server/src/wire/out"
)

func HealthToJSON(health domain.Health) out.Health {
	return out.Health{
		Status: health.Status,
		Time:   health.Time,
	}
}
