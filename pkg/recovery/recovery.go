package recovery

import (
	"io"

	"tp1/pkg/amqp"
	"tp1/pkg/logs"
	"tp1/pkg/sequence"
	ioutils "tp1/pkg/utils/io"
)

const filePath = "recovery.csv"

type Handler struct {
	file *ioutils.File
}

// NewHandler creates a new recovery handler.
func NewHandler() (*Handler, error) {
	file, err := ioutils.NewFile(filePath)
	if err != nil {
		return nil, err
	}

	return &Handler{
		file: file,
	}, nil
}

// Recover reads each line of the underlying file, parses it, and sends it through a Record channel.
// Upon failure when reading or parsing, the line gets skipped.
func (h *Handler) Recover(ch chan<- Record) {
	defer close(ch)

	for {
		line, err := h.file.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			logs.Logger.Errorf("Failed to read line: %v", err)
			continue
		}

		header, err := amqp.HeaderFromStrings(line)
		if err != nil {
			logs.Logger.Errorf("failed to recover header: %s", err.Error())
			continue
		}

		sequenceIds, err := sequence.DstsFromStrings(line[amqp.HeaderLen : len(line)-1])
		if err != nil {
			logs.Logger.Errorf("failed to recover sequence: %s", err.Error())
			continue
		}

		ch <- NewRecord(
			*header,
			sequenceIds,
			[]byte(line[len(line)-1]),
		)
	}
}

// Log saves a record into the underlying file.
func (h *Handler) Log(record Record) error {
	return h.file.Write(record.toString())
}

// Close closes the file descriptor linked to the underlying file.
func (h *Handler) Close() {
	h.file.Close()
}
