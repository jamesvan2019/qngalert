package qng

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestTimeParse(t *testing.T) {
	t1, err := time.Parse("2006-01-02T15:04:05+08:00", "2023-03-15T14:04:09+08:00")
	if err != nil {
		log.Println("timestamp parse error", err)
		return
	}
	fmt.Println(t1.Unix() - 8*3600)
}
