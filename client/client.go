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
	"golang.org/x/sync/errgroup"
)

const (
	pubStream = "MSG-PUB" // public messages stream
	dirStream = "MSG-DIR" // direct messages stream

	pubSubject = "MSG.ALL"
	dirSubject = "MSG.ALL.%s"
)

type Client struct {
	cfg *config.Config
	nc  *nats.Conn
	js  jetstream.JetStream
}

func NewClient(cfg *config.Config) (*Client, error) {
	nc, err := nats.Connect(cfg.NatsServerURL)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	js, err := jetstream.NewWithDomain(nc, cfg.NatsHubDomain)
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
	if msg.Text == "" {
		return entity.ErrInvalidArguments
	}

	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if len(msg.To) > 0 {
		return c.nc.Publish(fmt.Sprintf(dirSubject, msg.To), data)
	}
	return c.nc.Publish(pubSubject, data)
}

func (c *Client) PullMessages() ([]*entity.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pubMsgs []*entity.Message
	var dirMsgs []*entity.Message
	group, gctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		consumer, err := c.upsertPublicConsumer(gctx)
		if err != nil {
			return fmt.Errorf("public: %w", err)
		}

		pubMsgs, err = c.pullMessages(consumer)
		if err != nil {
			return fmt.Errorf("public messages: %w", err)
		}

		return nil
	})
	group.Go(func() error {
		consumer, err := c.upsertDirectConsumer(gctx)
		if err != nil {
			return fmt.Errorf("direct: %w", err)
		}

		dirMsgs, err = c.pullMessages(consumer)
		if err != nil {
			return fmt.Errorf("direct messages: %w", err)
		}

		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}

	return append(pubMsgs, dirMsgs...), nil
}

func (c *Client) pullMessages(consumer jetstream.Consumer) ([]*entity.Message, error) {
	const batchSize = 16

	batch, err := consumer.FetchNoWait(batchSize)
	if err != nil || batch.Error() != nil {
		return nil, errors.Join(err, fmt.Errorf("batch: %w", batch.Error()))
	}

	var msgs []*entity.Message
	for m := range batch.Messages() {
		var msg entity.Message
		if err := msg.Unmarshal(m.Data()); err != nil {
			return nil, m.TermWithReason(fmt.Errorf("unmarshal: %w", err).Error())
		}

		msgs = append(msgs, &msg)
		m.Ack()
	}

	return msgs, nil
}

func (c *Client) upsertPublicConsumer(ctx context.Context) (jetstream.Consumer, error) {
	stream, err := c.js.Stream(ctx, pubStream)
	if err != nil {
		return nil, fmt.Errorf("stream: %w", err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx,
		jetstream.ConsumerConfig{
			Durable:           c.pubConsumerName(),
			FilterSubject:     pubSubject,
			DeliverPolicy:     jetstream.DeliverLastPolicy,
			MemoryStorage:     true,
			InactiveThreshold: time.Hour,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("consumer: %w", err)
	}

	return consumer, nil
}

func (c *Client) upsertDirectConsumer(ctx context.Context) (jetstream.Consumer, error) {
	stream, err := c.js.Stream(ctx, dirStream)
	if err != nil {
		return nil, fmt.Errorf("stream: %w", err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx,
		jetstream.ConsumerConfig{
			Name:          c.dirConsumerName(),
			FilterSubject: fmt.Sprintf(dirSubject, c.cfg.Credentials.Username),
			DeliverPolicy: jetstream.DeliverAllPolicy,
			MemoryStorage: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("consumer: %w", err)
	}

	return consumer, nil
}

func (c *Client) pubConsumerName() string {
	return fmt.Sprintf("%s-%s", pubStream, c.cfg.Credentials.Username)
}

func (c *Client) dirConsumerName() string {
	return fmt.Sprintf("%s-%s", dirStream, c.cfg.Credentials.Username)
}

func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.js.DeleteConsumer(ctx, pubStream, c.pubConsumerName())
	c.js.DeleteConsumer(ctx, dirStream, c.dirConsumerName())

	c.nc.Close()

	return nil
}
