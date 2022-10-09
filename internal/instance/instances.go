package instance

import (
	"github.com/JoachimFlottorp/Linnea/internal/redis"
	"github.com/JoachimFlottorp/Linnea/internal/s3"
)

type InstanceList struct {
	Redis   redis.Instance
	Storage s3.Instance
}
