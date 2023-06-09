package qng

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestTimeParse(t *testing.T) {
	t1, err := time.Parse("2006-01-02T15:04:05", "2023-03-16T09:30:26")
	if err != nil {
		log.Println("timestamp parse error", err)
		return
	}
	fmt.Println(t1.Unix(), time.Now().Unix())
}
