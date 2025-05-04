package kafka

import (
	"context"

	"github.com/pkg/errors"
)

type writingDataWithValue interface {
	GetKafkaValue(context.Context) ([]byte, error)
	GetKafkaKey(context.Context) (string, error)
}

type writingData struct {
	Value interface{}
	Key   string
}

type kafkaClient struct {
}

func NewKafka()

func getWritingData(ctx context.Context, data interface{}) (writingData, error) {
	switch ent := data.(type) {
	case writingData:
		return ent, nil
	}

	var err error
	w := writingData{}
	valueData, ok := data.(writingDataWithValue)
	if !ok {
		return writingData{}, errors.New("have not func GetKafkaValue")
	}

	w.Value, err = valueData.GetKafkaValue(ctx)
	if err != nil {
		return w, err
	}

	w.Key, err = valueData.GetKafkaKey(ctx)
	if err != nil {
		return w, err
	}

	return w, nil
}

func () Write(ctx context.Context, topic string, data interface{}) error {
	writingData, err := getWritingData(ctx, data)
	if err != nil {
		return err
	}

}
