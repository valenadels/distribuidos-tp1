package top_n_playtime

import (
	"tp1/internal/errors"
	"tp1/internal/worker"
	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/message"
)

type filter struct {
	w *worker.Worker
	n uint8
	heap  message.MinHeapPlaytime
}

func New() (worker.Filter, error) {
	w, err := worker.New()
	if err != nil {
		return nil, err
	}
	return &filter{w: w}, nil
}

func (f *filter) Init() error {
	f.heap = message.MinHeapPlaytime{}
	return f.w.Init()
}

func (f *filter) Start() {
	f.n = uint8(f.w.Query.(float64))
	f.w.Start(f)
}

func (f *filter) Process(delivery amqp.Delivery) {
	messageId := message.ID(delivery.Headers[amqp.MessageIdHeader].(uint8))

	if messageId == message.EofMsg {

		f.publish()
		f.heap = nil

		if err := f.w.Broker.HandleEofMessage(f.w.Id, f.w.Peers, delivery.Body, nil, f.w.InputEof, f.w.OutputsEof...); err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
		}

		
	} else if messageId == message.GameWithPlaytimeID {
		msg, err := message.DateFilteredReleasesFromBytes(delivery.Body)
		if err != nil {
			logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
			return
		}
		
		// update heap with new data
		f.heap.UpdateReleases(msg,int(f.n))		
	} else {
		logs.Logger.Errorf(errors.InvalidMessageId.Error(), messageId)
	}
}

func (f *filter) publish() {

	if f.heap == nil {
		return
	}

	topNPlaytime := message.ToTopNPlaytimeMessage(f.n,&f.heap) 
	b, err := topNPlaytime.ToBytes()
	if err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToParse.Error(), err.Error())
	}

	headers := map[string]any{amqp.MessageIdHeader: message.GameWithPlaytimeID}
	if err = f.w.Broker.Publish(f.w.Outputs[0].Exchange, f.w.Outputs[0].Key, b, headers); err != nil {
		logs.Logger.Errorf("%s: %s", errors.FailedToPublish.Error(), err)
	}
}
