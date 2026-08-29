package broker

import "errors"

var (
	ErrNoBrokers     = errors.New("no brokers were provided")
	ErrTopicRequired = errors.New("topic is required")
)
