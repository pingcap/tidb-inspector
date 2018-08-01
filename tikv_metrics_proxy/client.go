// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"sync"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newConn(addr string, security Security) (*grpc.ClientConn, error) {
	opt := grpc.WithInsecure()
	if len(security.ClusterSSLCA) != 0 {
		tlsConfig, err := security.ToTLSConfig()
		if err != nil {
			return nil, errors.Trace(err)
		}
		opt = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}
	conn, err := grpc.Dial(addr, opt)
	if err != nil {
		log.Errorf("tikv store '%s', grpc dial error, %v", addr, err)
		return nil, errors.Trace(err)
	}
	return conn, nil
}

type rpcClient struct {
	sync.RWMutex
	isClosed bool
	conns    map[string]*grpc.ClientConn
	security Security
}

func newRPCClient(security Security) *rpcClient {
	return &rpcClient{
		conns:    make(map[string]*grpc.ClientConn),
		security: security,
	}
}

func (c *rpcClient) getConn(addr string) (*grpc.ClientConn, error) {
	c.RLock()
	if c.isClosed {
		c.RUnlock()
		return nil, errors.Errorf("rpcClient is closed")
	}
	conn, ok := c.conns[addr]
	c.RUnlock()
	if !ok {
		var err error
		conn, err = c.createConn(addr)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (c *rpcClient) createConn(addr string) (*grpc.ClientConn, error) {
	c.Lock()
	if c.isClosed {
		c.Unlock()
		return nil, errors.Errorf("rpcClient is closed")
	}
	defer c.Unlock()
	conn, err := newConn(addr, c.security)
	if err != nil {
		return nil, err
	}
	c.conns[addr] = conn
	return conn, nil
}

func (c *rpcClient) closeConns() {
	c.Lock()
	if !c.isClosed {
		c.isClosed = true
		for _, conn := range c.conns {
			conn.Close()
		}
		c.conns = nil
	}
	c.Unlock()
}
