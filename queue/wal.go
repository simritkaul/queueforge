package queue

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

type WAL struct {
	file *os.File;
	writer *bufio.Writer;
}


type WALRecord struct {
	Type uint8;
	Data []byte;
}


const (
	RecordEnqueue uint8 = 1
	RecordDequeue uint8 = 2
	RecordAck     uint8 = 3
)

/*
This creates a file that's readable and writable by the owner (6), 
and readable only by group and others (4 each).
*/
const defaultFilePerms = 0644;

// Opens a Write-Ahead Log
func OpenWAL (path string) (*WAL, error) {
	file, err := os.OpenFile(
		path, 
		os.O_CREATE | os.O_RDWR | os.O_APPEND,
		defaultFilePerms,
	)

	if err != nil {
		return nil, err;
	}

	return &WAL{
		file: file,
		writer: bufio.NewWriter(file),
	}, nil;
}

// Closes a Write-Ahead Log
func (w *WAL) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err;
	}

	return w.file.Close();
}

// Appends a record to the WAL
func (w *WAL) Append(recordType uint8, data []byte) error {
	// Write record type
	if err := w.writer.WriteByte(recordType); err != nil {
		return err;
	}

	// Write data length
	if err := binary.Write(w.writer, binary.BigEndian, uint32(len(data))); err != nil {
		return err;
	}

	// Write payload
	if _, err := w.writer.Write(data); err != nil {
		return err;
	}

	// Durability guarantee
	if err := w.writer.Flush(); err != nil {
		return err;
	}

	return w.file.Sync();
}

func (w *WAL) Replay (handler func(WALRecord) error) error {
	if _, err := w.file.Seek(0, 0); err != nil {
		return err;
	}

	reader := bufio.NewReader(w.file);

	for {
		recordType, err := reader.ReadByte();
		if err == io.EOF {
			return nil;
		}

		if err != nil {
			return err;
		}

		var size uint32;

		if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
			// partial record -> stop replay
			return nil;
		}

		data := make([]byte, size);
		if _, err := io.ReadFull(reader, data); err != nil {
			// partial record -> stop replay
			return nil;
		}

		if err := handler(WALRecord{
			Type: recordType,
			Data: data,
		}); err != nil {
			return err;
		}
	}
}