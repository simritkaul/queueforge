package internal

import (
	"bytes"
	"encoding/binary"
)

func DecodeEnqueue (data []byte) (string, string, error) {
	reader := bytes.NewReader(data);

	var idLen uint32;

	if err := binary.Read(reader, binary.BigEndian, &idLen); err != nil {
		return "", "", err;
	}

	jobId := make([]byte, idLen);

	if _, err := reader.Read(jobId); err != nil {
		return "", "", err;
	} 

	payload := make([]byte, reader.Len());

	if _, err := reader.Read(payload); err != nil {
		return "", "", err;
	}

	return string(jobId), string(payload), nil;
}

func EncodeEnqueue (jobId string, payload string) ([]byte, error) {
	var buf bytes.Buffer;

	if err := binary.Write(&buf, binary.BigEndian, uint32(len(jobId))); err != nil {
		return nil, err;
	}

	if _, err := buf.Write([]byte(jobId)); err != nil {
		return nil, err;
	}

	if _, err := buf.Write([]byte(payload)); err != nil {
		return nil, err;
	}

	return buf.Bytes(), nil;
}