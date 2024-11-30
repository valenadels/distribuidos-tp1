package gateway

import (
	"strconv"
	"strings"
	"sync"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type ChunkSender struct {
	id           int //0 for games, 1 for reviews
	channel      <-chan ChunkItem
	broker       amqp.MessageBroker
	exchange     string
	maxChunkSize uint8
	chunks       map[string][]any
	routingKey   string
}

type ChunkItem struct {
	Msg      any //DataCSVGames or DataCSVReviews
	ClientId string
	BatchNum uint32
}

func newChunkSender(id int, channel <-chan ChunkItem, broker amqp.MessageBroker, exchange string, chunkMaxSize uint8, routingKey string) *ChunkSender {
	return &ChunkSender{
		id:           id,
		channel:      channel,
		broker:       broker,
		exchange:     exchange,
		chunks:       make(map[string][]any),
		maxChunkSize: chunkMaxSize,
		routingKey:   routingKey,
	}
}

func startChunkSender(id int, clientAckChannels *sync.Map, channel <-chan ChunkItem, broker amqp.MessageBroker, exchange string, chunkMaxSize uint8, routingKey string) {
	s := newChunkSender(id, channel, broker, exchange, chunkMaxSize, routingKey)
	for {
		item := <-channel
		s.updateChunk(clientAckChannels, item, item.Msg == nil)
	}
}

func (s *ChunkSender) updateChunk(clientAckChannels *sync.Map, item ChunkItem, eof bool) {
	clientId := item.ClientId

	if !eof {
		if _, ok := s.chunks[clientId]; !ok {
			s.chunks[clientId] = make([]any, 0, s.maxChunkSize)
		}
		s.chunks[clientId] = append(s.chunks[clientId], item)
	}

	s.sendChunk(clientAckChannels, eof, clientId, item.BatchNum)
}

// sendChunk sends a chunks of data to the broker if the chunks is full or the eof flag is true
// In case eof is true, it sends an EOF message to the broker
// if a chunks was sent restarts count
func (s *ChunkSender) sendChunk(clientAckChannels *sync.Map, eof bool, clientId string, batchNum uint32) {
	messageId := matchMessageId(s.id)
	chunk := s.chunks[clientId]
	ackSent := false

	if len(chunk) >= int(s.maxChunkSize) || eof {
		bytes, err := toBytes(messageId, chunk)
		if err != nil {
			logs.Logger.Errorf("Error converting chunks to bytes: %s", err.Error())
		}

		sequenceId := strings.Replace(clientId, "-", "", -1) + "-" + strconv.FormatUint(uint64(batchNum), 10)
		headers := amqp.Header{MessageId: messageId, ClientId: clientId, SequenceId: sequenceId}
		err = s.broker.Publish(s.exchange, s.routingKey, bytes, headers)
		if err != nil {
			logs.Logger.Errorf("Error publishing chunks: %s", err.Error())
		}

		s.chunks[clientId] = make([]any, 0, s.maxChunkSize)
		sendAckThroughChannel(clientAckChannels, clientId)
		ackSent = true

	}

	if eof {
		sequenceId := strings.Replace(clientId, "-", "", -1) + "-" + strconv.FormatUint(uint64(batchNum+1), 10)
		headers := amqp.Header{MessageId: message.EofMsg, ClientId: clientId, SequenceId: sequenceId}
		err := s.broker.Publish(s.exchange, s.routingKey, amqp.EmptyEof, headers)
		if err != nil {
			logs.Logger.Errorf("Error publishing EOF: %s", err.Error())
		}
		logs.Logger.Infof("Sent eof with key %v", s.routingKey)
		delete(s.chunks, clientId)
		if !ackSent {
			sendAckThroughChannel(clientAckChannels, clientId)
		}
	}

}

func toBytes(msgId message.ID, chunk []any) ([]byte, error) {
	if msgId == message.ReviewIdMsg {
		reviews := make([]message.DataCSVReviews, 0, len(chunk))
		for _, v := range chunk {
			reviews = append(reviews, v.(ChunkItem).Msg.(message.DataCSVReviews))
		}
		return message.ReviewsFromClientReviews(reviews)
	}
	games := make([]message.DataCSVGames, 0, len(chunk))
	for _, v := range chunk {
		games = append(games, v.(ChunkItem).Msg.(message.DataCSVGames))
	}
	return message.GamesFromClientGames(games)
}

func sendAckThroughChannel(clientAckChannels *sync.Map, clientID string) {
	if clientChanI, exists := clientAckChannels.Load(clientID); exists {
		clientChan := clientChanI.(chan []byte)
		clientChan <- []byte{0x01}
	} else {
		logs.Logger.Errorf("No client channel found for clientID %v", clientID)
	}
}
