package main

import (
  "github.com/axetroy/generic-pool"
)

type Connection struct {
  online bool
}

func (c *Connection) Connect() (err error) {
  return
}

func (c *Connection) send(data []byte) {

}

func (c *Connection) Close() (err error) {
  return
}

func main() {

  p := pool.New(pool.Config{
    Creator: func(p *pool.Pool) (interface{}, error) {
      // create connection
      connection := Connection{
        online: true,
      }

      // connect
      if err := connection.Connect(); err != nil {
        return nil, err
      }

      // return connection
      return connection, nil
    },
    Destroyer: func(p *pool.Pool, resource interface{}) (err error) {
      // parse the connection
      connection := resource.(Connection)

      return connection.Close()
    },
  }, pool.Options{Min: 5, Max: 50, Idle: 60})

  if resource, err := p.Get(); err != nil {
    panic(err)
  } else {

    connection := resource.(Connection)

    defer func() {
      //connection.Close()
      // You don't need to close by manual, resource pool will do this
    }()

    connection.send([]byte("Hello world"))
  }

}
