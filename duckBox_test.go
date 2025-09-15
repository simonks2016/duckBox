package DuckBox

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"testing"

	"github.com/nsqio/go-nsq"
	"github.com/simonks2016/duckBox/cmd/producer"
	"github.com/simonks2016/duckBox/cmd/worker"
)

func TestDuckBox(t *testing.T) {

	ctx := context.Background()
	pool := worker.NewNSQConsumerPool([]string{"http://127.0.0.1:4161"}, worker.Nsqd{
		IsHttp:  false,
		Address: "127.0.0.1:4150",
	})

	pool.AddHandler(&worker.TopicPolicy{
		Topic:      "cmd.comment.create",
		Channel:    "worker002",
		DLQTopic:   "comment.dlq",
		UseLookupd: false,
		NSQd: []worker.Nsqd{{
			IsHttp:  false,
			Address: "127.0.0.1:4150"},
		},
		Handler: func(ctx context.Context, msg *nsq.Message) error {

			pool = worker.GetPool(ctx)
			fmt.Println(pool)
			fmt.Printf("%s\n", string(msg.Body))
			return nil
		},
	})

	producerPool, err := producer.NewNSQPool("127.0.0.1", 100)
	if err != nil {
		t.Fatal(err)
		return
	}
	err = pool.Start(ctx, producerPool)
	if err != nil {
		t.Fatal(err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

}
