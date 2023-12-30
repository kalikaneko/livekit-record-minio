package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	lksdk "github.com/livekit/server-sdk-go"
)

var (
	defaultLayout = "speaker"
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

func newRecordingName(room, shareWith string) string {
	return fmt.Sprintf("recording-%s-%s-%s.ogg", room, randomToken(8), shareWith)
}

func handleRecordStart(e echo.Context) error {
	queryParams := c.QueryParams()

	room := queryParams.Get("room")
	shareWith := queryParams.Get("shareWith")

	var res result

	rec, err := startRecording(room, shareWith)
	if err != nil {
		log.Println("error:", err.Error())
		res = result{Error: "cannot start recording, check logs"}
	} else {
		// this allows us to avoid other users from stopping the recording
		rec.StartedBy = user
		liveRecordings = append(liveRecordings, rec)
		log.Println("Started recording:", rec.FileName)
		res = result{Message: "ok"}
	}

	return c.JSON(http.StatusOK, res)

func startRecording(room, shareWith string) (*recording, error) {

	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitSecretKey := os.Getenv("LIVEKIT_API_SECRET")
	livekitInstance := os.Getenv("LIVEKIT_INSTANCE")
	livekitURL := fmt.Sprintf("https://%s", livekitInstance)

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Key := os.Getenv("S3_KEY")
	s3Secret := os.Getenv("S3_SECRET")

	if apiKey == "" || secretKey == "" {
		log.Fatal("missing LK_API_KEY or LK_API_SECRET")
	}

	egressClient := lksdk.NewEgressClient(livekitURL, livekitAPIKey, livekitSecretKey)
	fileName := newRecordingName(room, shareWith)
	fileRequest := &livekit.RoomCompositeEgressRequest{
		RoomName:  room,
		Layout:    "speaker",
		AudioOnly: true,
		Output: &livekit.RoomCompositeEgressRequest_File{
			File: &livekit.EncodedFileOutput{
				FileType: livekit.EncodedFileType_MP4,
				Filepath: fileName,
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey: s3Key,
						Secret:    s3Secret,
						Region:    s3Endpoint,
						Bucket:    s3Bucekt,
					},
				},
			},
		},
		// uncomment to use your own templates
		// CustomBaseUrl: "https://my-custom-template.com",
	}

	ctx := context.Background()
	info, err := egressClient.StartRoomCompositeEgress(ctx, fileRequest)
	if err != nil {
		return nil, err
	}

	rec := &recording{
		EgressID:     info.EgressID,
		EgressClient: egressClient,
		FileName:     fileName,
		Room:         room,
		Started:      time.Now(),
	}
	return rec, nil
}

func stopRecording(room string) error {
	livekitInstance := os.Getenv("LIVEKIT_INSTANCE")
	livekitURL := fmt.Sprintf("https://%s", livekitInstance)
	apiKey := os.Getenv("LK_API_KEY")
	secretKey := os.Getenv("LK_API_SECRET")

	if apiKey == "" || secretKey == "" {
		log.Fatal("missing LK_API_KEY or LK_API_SECRET")
	}

	ctx := context.Background()
	for _, rec := range liveRecordings {
		if rec.Room == room {
			log.Println("trying to stop", rec.EgressID)

			egressClient := lksdk.NewEgressClient(
				livekitURL, apiKey, secretKey)

			_, err := egressClient.StopEgress(ctx, &livekit.StopEgressRequest{
				EgressId: rec.EgressID,
			})
			if err != nil {
				log.Println(err)
			}
			return nil
		}
	}
	return errors.New("cannot find recording for room:" + room)
}

func unlistRecordingForRoom(room string) {
	for i, rec := range liveRecordings {
		if rec.Room == room {
			log.Println("remove live recording for room", room)
			liveRecordings = append(liveRecordings[:i], liveRecordings[i+1:]...)
		}
	}
}

func randomToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz1234567890"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
