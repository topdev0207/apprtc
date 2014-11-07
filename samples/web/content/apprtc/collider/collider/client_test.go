// Copyright (c) 2014 The WebRTC project authors. All Rights Reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package collider

import (
	"collidertest"
	"testing"
)

func TestNewClient(t *testing.T) {
	id := "abc"
	c := newClient(id)
	if c.id != id {
		t.Errorf("newClient(%q).id = %s, want %q", id, c.id, id)
	}
	if c.rwc != nil {
		t.Errorf("newClient(%q).rwc = %v, want nil", id, c.rwc)
	}
	if c.msgs != nil {
		t.Errorf("newClient(%q).msgs = %v, want nil", id, c.msgs)
	}
}

// Tests that registering the client twice will fail.
func TestClientRegister(t *testing.T) {
	id := "abc"
	c := newClient(id)
	var rwc collidertest.MockReadWriteCloser
	if err := c.register(&rwc); err != nil {
		t.Errorf("newClient(%q).register(%v) got error: %s, want nil", id, &rwc, err.Error())
	}
	if c.rwc != &rwc {
		t.Errorf("client.rwc after client.register(%v) = %v, want %v", &rwc, c.rwc, &rwc)
	}

	// Register again and it should fail.
	if err := c.register(&rwc); err == nil {
		t.Errorf("Second call of client.register(%v): nil, want !nil error", &rwc)
	}
}

// Tests that queued messages are delivered in sendQueued.
func TestClientSendQueued(t *testing.T) {
	src := newClient("abc")
	src.enqueue("hello")

	dest := newClient("def")
	rwc := collidertest.MockReadWriteCloser{Closed: false}

	dest.register(&rwc)
	src.sendQueued(dest)

	if rwc.Msg == "" {
		t.Errorf("After sending queued messages from src to dest, dest.rwc.Msg = %v, want non-empty", rwc.Msg)
	}
	if len(src.msgs) != 0 {
		t.Errorf("After sending queued messages from src to dest, src.msgs = %v, want empty", src.msgs)
	}
}

// Tests that messages are queued when the other client is not registered, or delivered immediately otherwise.
func TestClientSend(t *testing.T) {
	src := newClient("abc")
	dest := newClient("def")

	// The message should be queued since dest has not registered.
	m := "hello"
	if err := src.send(dest, m); err != nil {
		t.Errorf("When dest is not registered, src.send(dest, %q) got error: %s, want nil", m, err.Error())
	}
	if len(src.msgs) != 1 || src.msgs[0] != m {
		t.Errorf("After src.send(dest, %q) when dest is not registered, src.msgs = %v, want [%q]", m, src.msgs, m)
	}

	rwc := collidertest.MockReadWriteCloser{Closed: false}
	dest.register(&rwc)

	// The message should be sent this time.
	m2 := "hi"
	src.send(dest, m2)

	if rwc.Msg == "" {
		t.Errorf("When dest is registered, after src.send(dest, %q), dest.rwc.Msg = %v, want %q", m2, rwc.Msg, m2)
	}
	if len(src.msgs) != 1 || src.msgs[0] != m {
		t.Errorf("When dest is registered, after src.send(dest, %q), src.msgs = %v, want [%q]", m2, src.msgs, m)
	}
}

// Tests that closing the client will close the ReadWriteCloser.
func TestClientClose(t *testing.T) {
	c := newClient("abc")
	rwc := collidertest.MockReadWriteCloser{Closed: false}

	c.register(&rwc)
	c.close()
	if !rwc.Closed {
		t.Errorf("After client.close(), rwc.Closed = %t, want true", rwc.Closed)
	}
}
