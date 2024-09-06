package client

import (
	"fmt"

	"github.com/cnaize/nth-chat/config"
	"github.com/cnaize/nth-chat/entity"
	"github.com/nats-io/nats.go"
)

const sendMsgSub = "msg.send.%s"

type Client struct {
	cfg     *config.Config
	nc      *nats.Conn
	pullCh  chan *nats.Msg
	pullSub *nats.Subscription
}

func NewClient(cfg *config.Config) (*Client, error) {
	nc, err := nats.Connect(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return &Client{
		cfg: cfg,
		nc:  nc,
	}, nil
}

func (c *Client) SendMessage(msg entity.Message) error {
	if msg.To == "" || msg.Text == "" {
		return fmt.Errorf("emtpy recipient or text")
	}

	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return c.nc.Publish(fmt.Sprintf(sendMsgSub, msg.To), data)
}

func (c *Client) PullMessages() ([]*entity.Message, error) {
	// TODO: CHANGE!!!
	if c.pullSub == nil {
		pullCh := make(chan *nats.Msg, 128)
		pullSub, err := c.nc.ChanSubscribe(fmt.Sprintf(sendMsgSub, c.cfg.Username), pullCh)
		if err != nil {
			return nil, fmt.Errorf("subscribe: %w", err)
		}

		c.pullCh = pullCh
		c.pullSub = pullSub
	}

	var res []*entity.Message
	err := func() error {
		for {
			select {
			case msg, ok := <-c.pullCh:
				if !ok {
					return fmt.Errorf("chan is closed")
				}

				var message entity.Message
				if err := message.Unmarshal(msg.Data); err != nil {
					fmt.Println("Error: unmarshal message:", err)
					continue
				}
				res = append(res, &message)
			default:
				return nil
			}
		}
	}()

	return res, err
}

func (c *Client) Close() error {
	err := c.pullSub.Unsubscribe()
	c.nc.Close()

	return err
}
