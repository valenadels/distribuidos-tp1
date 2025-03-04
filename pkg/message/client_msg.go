package message

import (
	"encoding/binary"
	"fmt"
	"net"
	"tp1/pkg/utils/io"
)

type ClientMessage struct {
	BatchNum uint32
	DataLen  uint32
	Data     []byte
}

type DataCSVGames struct {
	AppID                   int64
	Name                    string
	ReleaseDate             string
	EstimatedOwners         string
	PeakCCU                 int64
	RequiredAge             int64
	Price                   float64
	DiscountDLCCount        int64
	Blank                   int64
	AboutTheGame            string
	SupportedLanguages      string
	FullAudioLanguages      string
	Reviews                 string
	HeaderImage             string
	Website                 string
	SupportURL              string
	SupportEmail            string
	Windows                 bool
	Mac                     bool
	Linux                   bool
	MetacriticScore         int64
	MetacriticURL           string
	UserScore               int64
	Positive                int64
	Negative                int64
	ScoreRank               float64
	Achievements            int64
	Recommendations         int64
	Notes                   string
	AveragePlaytimeForever  int64
	AveragePlaytimeTwoWeeks int64
	MedianPlaytimeForever   int64
	MedianPlaytimeTwoWeeks  int64
	Developers              string
	Publishers              string
	Categories              string
	Genres                  string
	Tags                    string
	Screenshots             string
	Movies                  string
}

type DataCSVReviews struct {
	AppID       int64
	AppName     string
	ReviewText  string
	ReviewScore int64
	ReviewVotes int64
}

const lenBytes = 4

func SendMessage(conn net.Conn, msg ClientMessage) error {
	finalMessage := make([]byte, 0, lenBytes*2+len(msg.Data))
	batchNumBytes := make([]byte, lenBytes)
	binary.BigEndian.PutUint32(batchNumBytes, msg.BatchNum)
	finalMessage = append(finalMessage, batchNumBytes...)
	lenBytes := make([]byte, lenBytes)
	binary.BigEndian.PutUint32(lenBytes, msg.DataLen)
	finalMessage = append(finalMessage, lenBytes...)
	finalMessage = append(finalMessage, msg.Data...)
	if err := io.SendAll(conn, finalMessage); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}

func DataCSVReviewsFromBytes(b []byte) (DataCSVReviews, error) {
	var m DataCSVReviews
	return m, fromBytes(b, &m)
}

func (m DataCSVReviews) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func DataCSVGamesFromBytes(b []byte) (DataCSVGames, error) {
	var m DataCSVGames
	return m, fromBytes(b, &m)
}

func (m DataCSVGames) ToBytes() ([]byte, error) {
	return toBytes(m)
}
