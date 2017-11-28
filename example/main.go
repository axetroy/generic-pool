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

  pool.New(pool.Config{
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
}
