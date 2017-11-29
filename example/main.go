package main

import (
  "github.com/axetroy/generic-pool"
)

type faceConnection struct {
}

func (c *faceConnection) Connect() (err error) {
  return
}

func (c *faceConnection) send(data []byte) {

}

func (c *faceConnection) OnClose(func()) (err error) {
  return
}

func (c *faceConnection) Close() (err error) {
  return
}

func main() {

  p, _ := pool.New(pool.Config{
    Creator: func(p *pool.Pool, id int64) (interface{}, error) {
      // create an face connection
      faceConnection := faceConnection{}

      // connect
      if err := faceConnection.Connect(); err != nil {
        return nil, err
      }

      // when connection close by remote, we should remove it from pool
      faceConnection.OnClose(func() {
        // release the resource
        p.Release(id)
      })

      // return this
      return faceConnection, nil
    },
    Destroyer: func(p *pool.Pool, resource interface{}) (err error) {
      // parse the connection
      faceConnection := resource.(faceConnection)

      return faceConnection.Close()
    },
  }, pool.Options{Min: 5, Max: 50, Idle: 60})

  // Get the resource
  resource, err := p.Get()

  if err != nil {
    panic(err)
  }

  // parse the resource to connection
  faceConnection := resource.(faceConnection)

  defer func() {
    // faceConnection.Close()
    // You don't need to close by manual, resource pool will do this
  }()

  // send data
  faceConnection.send([]byte("Hello world"))

}
