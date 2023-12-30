package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
)

var liveRecordings []*recording

func initEnv() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	initEnv()
	e := echo.New()
	e.GET("/start", handleRecordStart)
	e.GET("/stop", handleRecordStop)
	e.Logger.Fatal(e.Start(":3000"))
}

func handleRecordStart(c echo.Context) error {
	queryParams := c.QueryParams()

	room := queryParams.Get("room")
	shareWith := queryParams.Get("shareWith")

	var res result

	rec, err := startRecording(room, shareWith)
	if err != nil {
		log.Println("error:", err.Error())
		res = result{Error: "cannot start recording, check logs"}
	} else {
		liveRecordings = append(liveRecordings, rec)
		log.Println("Started recording:", rec.FileName)
		res = result{Message: "ok"}
	}

	return c.JSON(http.StatusOK, res)
}

func handleRecordStop(c echo.Context) error {
	queryParams := c.QueryParams()
	room := queryParams.Get("room")

	log.Println("there are", len(liveRecordings), "live recordings")
	log.Println("stopping recording")

	err := stopRecording(room)
	var res result
	if err != nil {
		res = result{Error: err.Error()}
	} else {
		unlistRecordingForRoom(room)
		res = result{Message: "ok"}
	}

	return c.JSON(http.StatusOK, res)
}
