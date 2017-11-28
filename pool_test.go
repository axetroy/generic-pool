package pool

import (
  "testing"
  "time"
  "fmt"
)

type Engine struct {
  Id string
}

func TestPool_Pool(t *testing.T) {
  p := New(Config{
    // create connection
    Creator: func(p *Pool) (interface{}, error) {
      return &Engine{
        Id: "hello id",
      }, nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5})

  resource, _ := p.Get()

  d := resource.(*Engine)

  if len(p.pool) != 1 {
    t.Errorf("pool length not equal 1")
    return
  }

  if d.Id != "hello id" {
    t.Errorf("The id not equal")
    return
  }

  if err := p.Destroy(); err != nil {
    t.Errorf("Destroy pool should not throw an error")
    return
  } else {
    if len(p.pool) != 0 {
      t.Errorf("The pool after destroy should be 0")
      return
    }

    // after destroy, if apply .Get(), it should throw an error
    r, err := p.Get()

    if err == nil {
      t.Errorf("It should throw an error after pool destroy if you call .Get()")
      return
    }

    if r != nil {
      t.Errorf("The resouce should be nil")
      return
    }

  }
}

// The pool size is define, no matter how many time you get, it must be
func TestPool_PoolSize(t *testing.T) {
  p := New(Config{
    // create connection
    Creator: func(p *Pool) (interface{}, error) {
      return &Engine{
        Id: "hello id",
      }, nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5})

  for i := 0; i < 100; i++ {
    p.Get() // increase pool +1
    //fmt.Println(i, p.pool)
    if i <= 4 {
      if len(p.pool) != i+1 {
        t.Errorf("The pool size should be %v", i+1)
        return
      }
    } else {
      if len(p.pool) != 5 {
        t.Errorf("The max pool size is %v and now is %v", 5, i)
        return
      }
    }
    time.Sleep(time.Millisecond * 10)
  }

  // first resource always is latest resource
  r0 := p.pool[0]
  r1 := p.pool[4]

  if r0.LastUseAt.UnixNano() < r1.LastUseAt.UnixNano() {
    t.Errorf("The fist resource is not the latest!")
    return
  }
}

// if resource is not use, it should be release
func TestPool_PoolIdle(t *testing.T) {
  p := New(Config{
    // create connection
    Creator: func(p *Pool) (interface{}, error) {
      return &Engine{
        Id: "hello id",
      }, nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5, Idle: 5})

  // make pool fulled
  for i := 0; i < 100; i++ {
    p.Get()
  }

  if len(p.pool) != 5 {
    t.Errorf("The pool length should be 5")
  }

  ticker := time.NewTicker(time.Second * 1)
  go func() {
    for _ = range ticker.C {
      p.checkIdle()
      fmt.Println(p.pool)
    }
  }()

  time.Sleep(time.Second * 6)

  if len(p.pool) != 1 {
    t.Errorf("The pool after all should be reduce to 1")
    return
  }

  if p.pool[0].Idle != true {
    t.Errorf("The idle pool should mark with idel")
    return
  }
}
