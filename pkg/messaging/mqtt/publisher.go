// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"time"

	"github.com/MainfluxLabs/mainflux"
	"github.com/MainfluxLabs/mainflux/pkg/messaging"
	"github.com/MainfluxLabs/mainflux/pkg/messaging/nats"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
)

var errPublishTimeout = errors.New("failed to publish due to timeout reached")

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	client  mqtt.Client
	timeout time.Duration
}

// NewPublisher returns a new MQTT message publisher.
func NewPublisher(address string, timeout time.Duration) (messaging.Publisher, error) {
	client, err := newClient(address, "mqtt-publisher", timeout)
	if err != nil {
		return nil, err
	}

	ret := publisher{
		client:  client,
		timeout: timeout,
	}
	return ret, nil
}

func (pub publisher) Publish(conn *mainflux.ConnByKeyRes, msg messaging.Message) error {
	msg, _, err := nats.SetMessageProfile(conn, msg)
	if err != nil {
		return err
	}

	if !msg.Profile.Retention {
		return nil
	}

	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	token := pub.client.Publish(conn.ChannelID, qos, false, data)
	if token.Error() != nil {
		return token.Error()
	}
	ok := token.WaitTimeout(pub.timeout)
	if !ok {
		return errPublishTimeout
	}

	return token.Error()
}

func (pub publisher) Close() error {
	pub.client.Disconnect(uint(pub.timeout))
	return nil
}
