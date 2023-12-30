package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

var (
	defaultLayout = "speaker"
)

type result struct {
	Error   string `json:"error"`
	Message string `json:"msg"`
}

type recording struct {
	Room         string
	Started      time.Time
	ShareType    int
	FileName     string
	EgressID     string
	EgressClient *lksdk.EgressClient
}

func newRecordingName(room, shareWith string) string {
	return fmt.Sprintf("recordings/recording-%s-%s-%s.ogg", room, randomToken(8), shareWith)
}

func startRecording(room, shareWith string) (*recording, error) {
	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitSecretKey := os.Getenv("LIVEKIT_API_SECRET")
	livekitInstance := os.Getenv("LIVEKIT_INSTANCE")
	livekitURL := fmt.Sprintf("https://%s", livekitInstance)

	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Key := os.Getenv("S3_KEY")
	s3Secret := os.Getenv("S3_SECRET")

	if livekitAPIKey == "" || livekitSecretKey == "" {
		log.Fatal("missing LIVEKIT_API_KEY or LIVEKIT_API_SECRET")
	}

	egressClient := lksdk.NewEgressClient(livekitURL, livekitAPIKey, livekitSecretKey)
	fileName := newRecordingName(room, shareWith)
	fileRequest := &livekit.RoomCompositeEgressRequest{
		RoomName:  room,
		Layout:    defaultLayout,
		AudioOnly: true,
		Output: &livekit.RoomCompositeEgressRequest_File{
			File: &livekit.EncodedFileOutput{
				FileType: livekit.EncodedFileType_OGG,
				Filepath: fileName,
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey:      s3Key,
						Secret:         s3Secret,
						Endpoint:       s3Endpoint,
						Bucket:         s3Bucket,
						Region:         "any",
						ForcePathStyle: true,
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
		EgressID:     info.EgressId,
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
	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitSecretKey := os.Getenv("LIVEKIT_API_SECRET")

	if livekitAPIKey == "" || livekitSecretKey == "" {
		log.Fatal("missing LIVEKIT_API_KEY or LK_API_SECRET")
	}

	ctx := context.Background()
	for _, rec := range liveRecordings {
		if rec.Room == room {
			log.Println("trying to stop", rec.EgressID)

			egressClient := lksdk.NewEgressClient(
				livekitURL, livekitAPIKey, livekitSecretKey)

			_, err := egressClient.StopEgress(ctx, &livekit.StopEgressRequest{
				EgressId: rec.EgressID,
			})
			if err != nil {
				log.Println("error stopping egress")
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
