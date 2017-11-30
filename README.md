## Generic pool manager

[![Build Status](https://travis-ci.org/axetroy/generic-pool.svg?branch=master)](https://travis-ci.org/axetroy/generic-pool)
[![Coverage Status](https://coveralls.io/repos/github/axetroy/generic-pool/badge.svg?branch=master)](https://coveralls.io/github/axetroy/generic-pool?branch=master)
![License](https://img.shields.io/badge/license-Apache-green.svg)

Manage the resource pool, like connection...

## Features

- [x] Thread safety at use
- [x] Graceful to create/destroy the resource/connection
- [x] Easy to Use
- [x] 100% test cover

## Usage

```bash
go get -v github.com/axetroy/generic-pool
```

```go
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
```

## Contributing

[Contributing Guid](https://github.com/axetroy/generic-pool/blob/master/CONTRIBUTING.md)

## Contributors

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
| [<img src="https://avatars1.githubusercontent.com/u/9758711?v=3" width="100px;"/><br /><sub>Axetroy</sub>](http://axetroy.github.io)<br />[💻](https://github.com/axetroy/generic-pool/commits?author=axetroy) [🐛](https://github.com/axetroy/generic-pool/issues?q=author%3Aaxetroy) 🎨 |
| :---: |
<!-- ALL-CONTRIBUTORS-LIST:END -->

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Faxetroy%2Fgeneric-pool.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Faxetroy%2Fgeneric-pool?ref=badge_large)