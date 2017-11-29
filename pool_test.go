package pool

import (
  "testing"
  "time"
  "errors"
  "fmt"
)

type Engine struct {
  Id string
}

func TestPool_Pool(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return &Engine{
        Id: "hello id",
      }, nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5, Idle: 60})

  resource, _ := p.Get()

  d := resource.(*Engine)

  if len(p.Pool) != 1 {
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
    if len(p.Pool) != 0 {
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
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return &Engine{
        Id: "hello id",
      }, nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5, Idle: 60})

  for i := 0; i < 100; i++ {
    p.Get() // increase pool +1
    //fmt.Println(i, p.pool)
    if i <= 4 {
      if len(p.Pool) != i+1 {
        t.Errorf("The pool size should be %v", i+1)
        return
      }
    } else {
      if len(p.Pool) != 5 {
        t.Errorf("The max pool size is %v and now is %v", 5, i)
        return
      }
    }
    time.Sleep(time.Millisecond * 10)
  }

  // first resource always is latest resource
  r0 := p.Pool[0]
  r1 := p.Pool[4]

  if r0.LastUseAt.UnixNano() < r1.LastUseAt.UnixNano() {
    t.Errorf("The fist resource is not the latest!")
    return
  }
}

// if resource is not use, it should be release
func TestPool_PoolIdle(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return &Engine{
        Id: "hello id",
      }, nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5, Idle: 3})

  // make pool fulled
  for i := 0; i < 100; i++ {
    p.Get()
  }

  if len(p.Pool) != 5 {
    t.Errorf("The pool length should be 5 not %v", len(p.Pool))
  }

  ticker := time.NewTicker(time.Millisecond * 100)
  go func() {
    for _ = range ticker.C {
      p.checkIdle()
    }
  }()

  time.Sleep(time.Second * 5)

  fmt.Println(len(p.Pool))

  if len(p.Pool) != 1 {
    t.Errorf("The pool after all should be reduce to 1 not %v", len(p.Pool))
    return
  }

  if p.Pool[0].Idle != true {
    t.Errorf("The idle pool should mark with idel")
    return
  }
}

// if can not create resource
func TestPool_PoolCreatorThrowError(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return nil, errors.New("can not connect to db")
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 5, Idle: 3})

  resource, err := p.Get()

  if err == nil {
    t.Errorf("Create result should fail, and return an error")
    return
  }

  if resource != nil {
    t.Errorf("The resouce should be nil")
  }

}

// if can not destroy resource
func TestPool_PoolDestroyerThrowError(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return errors.New("destroy connection fail")
    },
  }, Options{Min: 1, Max: 5, Idle: 3})

  p.Get() // generate an resource

  // destroy the pool
  err := p.Destroy()

  if err == nil {
    t.Errorf("Destroyer result should fail, and return an error")
    return
  }

  if p.Destroyed == true {
    t.Errorf("The pool didn't be destroy at all")
    return
  }

}

// destroy without any exist resource
func TestPool_PoolDestroyWithoutResource(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return errors.New("destroy connection fail")
    },
  }, Options{Min: 1, Max: 5, Idle: 3})

  // destroy the pool
  err := p.Destroy()

  if err != nil {
    t.Errorf("There are no one resource so the don't need be destroy")
    return
  }

  if p.Destroyed == false {
    t.Errorf("The pool shoud be destroy")
    return
  }

}

// destroy without any exist resource
func TestPool_PoolIdleWhenDestroyerThrow(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return errors.New("destroy connection fail")
    },
  }, Options{Min: 1, Max: 5, Idle: 2})

  // make pool fulled
  for i := 0; i < 100; i++ {
    p.Get()
  }

  ticker := time.NewTicker(time.Millisecond * 100)
  go func() {
    for _ = range ticker.C {
      p.checkIdle()
    }
  }()

  time.Sleep(time.Second * 4)

  if len(p.Pool) != 5 {
    t.Errorf("The pool length should be 1 not %v", len(p.Pool))
    return
  }

  fmt.Println(p.Pool)

  // every resource's property Destroyed should be false
  for _, resource := range p.Pool {
    if resource.Destroyed == true {
      t.Errorf("Every one resource property .Destroyed should be false")
    }
  }

}

func TestPool_PoolDestroyLikeExpect(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 2, Max: 5, Idle: 30})

  // make pool fulled
  for i := 0; i < 100; i++ {
    p.Get()
  }

  err := p.Destroy()

  if err != nil {
    t.Errorf("Destroy should succes not thorw an error")
    return
  }

  if len(p.Pool) != 0 {
    t.Errorf("The pool length should be 0 not %v", len(p.Pool))
    return
  }
}

func TestPool_PoolDefaultOptions(t *testing.T) {
  p1, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: -1, Max: 5, Idle: 30})

  if p1.Options.Min != 0 {
    t.Errorf("The options of Mix should be 0 not %v", p1.Options.Min)
    return
  }

  p2, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: -1, Max: -1, Idle: 30})

  if p2.Options.Max != p2.Options.Min+1 {
    t.Errorf("The options of Max should be Min+1 not %v", p2.Options.Max)
    return
  }

  p3, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 100})

  if p3.Options.Idle != 30 {
    t.Errorf("The options of Idle should be 30 not %v", p3.Options.Idle)
    return
  }

  _, err := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 5, Max: 1, Idle: 30})

  if err == nil {
    t.Errorf("The options of Min>Max, it shoud be an error %v", err)
    return
  }

}

func TestPool_PoolInvalidConfig(t *testing.T) {
  _, err := New(Config{
    // create connection
    Creator: nil,
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: -1, Max: 5, Idle: 30})

  if err == nil {
    t.Errorf("Dit not set Creator %v", err)
    return
  }

  _, err = New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: nil,
  }, Options{Min: -1, Max: 5, Idle: 30})

  if err == nil {
    t.Errorf("Dit not set Destroyer %v", err)
    return
  }

}

func TestPool_PoolReleaseResource(t *testing.T) {
  p, _ := New(Config{
    // create connection
    Creator: func(p *Pool, id int64) (interface{}, error) {
      return "This is a connection", nil
    },
    // destroy connection
    Destroyer: func(p *Pool, connection interface{}) (error) {
      return nil
    },
  }, Options{Min: 1, Max: 100, Idle: 30})

  // fulled the pool
  for i := 0; i < p.Options.Max+1; i++ {
    p.Get()
  }

  // release the resource one by one

  for _, resource := range p.Pool {
    if err := p.Release(resource.Id); err != nil {
      t.Errorf("Release resource should success: %v", err)
      return
    }
  }

  if len(p.Pool) != 0 {
    t.Errorf("Now the pool should be empty, not length of %v", len(p.Pool))
    return
  }

}
