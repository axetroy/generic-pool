## Generic pool manager

[![Build Status](https://travis-ci.org/axetroy/generic-pool.svg?branch=master)](https://travis-ci.org/axetroy/generic-pool)
![License](https://img.shields.io/badge/license-Apache-green.svg)

Manage the resource pool, like connection...

## Usage

```bash
go get -v github.com/axetroy/generic-pool
```

```go
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

```

## Contributing

[Contributing Guid](https://github.com/axetroy/generic-pool/blob/master/CONTRIBUTING.md)

## Contributors

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
| [<img src="https://avatars1.githubusercontent.com/u/9758711?v=3" width="100px;"/><br /><sub>Axetroy</sub>](http://axetroy.github.io)<br />[💻](https://github.com/axetroy/generic-pool/commits?author=axetroy) [🐛](https://github.com/axetroy/generic-pool/issues?q=author%3Aaxetroy) 🎨 |
| :---: |
<!-- ALL-CONTRIBUTORS-LIST:END -->

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Faxetroy%2Fnid.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Faxetroy%2Fnid?ref=badge_large)