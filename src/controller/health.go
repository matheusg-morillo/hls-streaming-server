package controller

import (
	"matflix/hls-streaming-server/src/domain"
	"time"
)

func Health() domain.Health {
	return domain.Health{
		Status: "Healthy",
		Time:   time.Now().Format(time.RFC3339),
	}
}
