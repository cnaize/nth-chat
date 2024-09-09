package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cnaize/nth-chat/config"
	"github.com/cnaize/nth-chat/entity"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	msgsBoxStream  = "MSGSBOX"
	msgsBoxSubject = "MSGS.BOX.%s"
)

type Client struct {
	cfg *config.Config
	nc  *nats.Conn
	js  jetstream.JetStream
}

func NewClient(cfg *config.Config) (*Client, error) {
	nc, err := nats.Connect(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("new jetstream: %w", err)
	}

	return &Client{
		cfg: cfg,
		nc:  nc,
		js:  js,
	}, nil
}

func (c *Client) SendMessage(msg entity.Message) error {
	if msg.To == "" || msg.Text == "" {
		return entity.ErrInvalidArguments
	}

	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return c.nc.Publish(fmt.Sprintf(msgsBoxSubject, msg.To), data)
}

func (c *Client) PullMessages() ([]*entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := c.js.Stream(ctx, msgsBoxStream)
	if err != nil {
		return nil, fmt.Errorf("stream: %w", err)
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx,
		jetstream.ConsumerConfig{
			Name:          fmt.Sprintf("%s-%s", msgsBoxStream, c.cfg.Creds.Username),
			FilterSubject: fmt.Sprintf(msgsBoxSubject, c.cfg.Creds.Username),
			DeliverPolicy: jetstream.DeliverAllPolicy,
			MemoryStorage: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("new consumer: %w", err)
	}

	batch, err := cons.FetchNoWait(32)
	if err != nil || batch.Error() != nil {
		return nil, errors.Join(err, fmt.Errorf("batch: %w", batch.Error()))
	}

	var msgs []*entity.Message
	for msg := range batch.Messages() {
		var emsg entity.Message
		if err := emsg.Unmarshal(msg.Data()); err != nil {
			return nil, msg.TermWithReason(fmt.Errorf("unmarshal: %w", err).Error())
		}

		msgs = append(msgs, &emsg)
		msg.DoubleAck(ctx)
	}

	return msgs, nil
}

func (c *Client) Close() error {
	c.nc.Close()
	return nil
}
