package main

import (
	"time"

	"github.com/labstack/echo"
	lksdk "github.com/livekit/server-sdk-go"
)

type recording struct {
	Room         string
	Started      time.Time
	StartedBy    string
	ShareType    int
	FileName     string
	EgressID     string
	EgressClient *lksdk.EgressClient
}

var liveRecordings []*recording

func main() {
	var liveRecordings []*recording

	e := echo.New()
	e.GET("/start", handleRecordStart)
	e.GET("/stop", handleRecordStop)
	e.Logger.Fatal(e.Start(":3000"))
}
