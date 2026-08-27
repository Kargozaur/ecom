package main

import "google.golang.org/grpc"

type Conns struct {
	userConn  *grpc.ClientConn
	orderConn *grpc.ClientConn
}

func (c *Conns) Close() {
	if c.orderConn != nil {
		c.orderConn.Close()
	}
	if c.userConn != nil {
		c.userConn.Close()
	}
}
