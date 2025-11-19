package pool_test

import (
	"fmt"

	"github.com/EshkinKot1980/metrics/internal/pool"
)

type Connection struct {
	Buffer []byte
	Socket string
}

func (c *Connection) Reset() {
	c.Buffer = c.Buffer[:0]
	c.Socket = ""
}

func ExamplePool() {
	connectionPool := pool.New(func() *Connection {
		return &Connection{
			Buffer: make([]byte, 0, 512),
		}
	})

	conn := connectionPool.Get()
	fmt.Printf("New connection - Socket: %s, Buffer size: %d\n", conn.Socket, len(conn.Buffer))

	conn.Socket = "127.0.0.1:8080"
	conn.Buffer = append(conn.Buffer, []byte("some data")...)
	fmt.Printf("After use - Socket: %s, Buffer size: %d\n", conn.Socket, len(conn.Buffer))

	connectionPool.Put(conn)
	_ = connectionPool.Get()

	fmt.Printf("After get (reuse connection) - Socket: %s, Buffer size: %d\n", conn.Socket, len(conn.Buffer))

	// Output:
	// New connection - Socket: , Buffer size: 0
	// After use - Socket: 127.0.0.1:8080, Buffer size: 9
	// After get (reuse connection) - Socket: , Buffer size: 0
}
