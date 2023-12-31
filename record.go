package livekitminio

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

var (
	defaultLayout = "speaker"
)

type Session struct {
	liveRecordings []*Recording
	mu             *sync.Mutex
}

func NewSession() *Session {
	return &Session{
		liveRecordings: make([]*Recording, 0),
		mu:             &sync.Mutex{},
	}
}

type Recording struct {
	Room         string
	Started      time.Time
	ShareType    int
	FileName     string
	EgressID     string
	EgressClient *lksdk.EgressClient
}

func (s *Session) StartRecording(room, shareWith string) (*Recording, error) {
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
	}

	ctx := context.Background()
	info, err := egressClient.StartRoomCompositeEgress(ctx, fileRequest)
	if err != nil {
		return nil, err
	}

	rec := &Recording{
		EgressID:     info.EgressId,
		EgressClient: egressClient,
		FileName:     fileName,
		Room:         room,
		Started:      time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.liveRecordings = append(s.liveRecordings, rec)
	log.Println("Started recording:", rec.FileName)

	return rec, nil
}

func (s *Session) StopRecording(room string) error {
	livekitInstance := os.Getenv("LIVEKIT_INSTANCE")
	livekitURL := fmt.Sprintf("https://%s", livekitInstance)
	livekitAPIKey := os.Getenv("LIVEKIT_API_KEY")
	livekitSecretKey := os.Getenv("LIVEKIT_API_SECRET")

	if livekitAPIKey == "" || livekitSecretKey == "" {
		log.Fatal("missing LIVEKIT_API_KEY or LK_API_SECRET")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	log.Println("there are", len(s.liveRecordings), "live recordings")
	for _, rec := range s.liveRecordings {
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

func (s *Session) UnlistRecordingForRoom(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, rec := range s.liveRecordings {
		if rec.Room == room {
			log.Println("remove live recording for room", room)
			s.liveRecordings = append(s.liveRecordings[:i], s.liveRecordings[i+1:]...)
		}
	}
}

func newRecordingName(room, shareWith string) string {
	return fmt.Sprintf("recordings/recording-%s-%s-%s.ogg", room, randomToken(8), shareWith)
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
