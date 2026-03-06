package utils

import (
	"strconv"
	"time"
)

func GetTimeString() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
