package pool

import (
  "errors"
  "time"
  "sort"
)

type CreatorFunc func(p *Pool, id int64) (resource interface{}, err error)
type DestroyerFunc func(p *Pool, resource interface{}) (err error)

type Pool struct {
  Config    Config
  Options   Options
  Pool      []*Resource
  Destroyed bool
  index     int64
}

type Resource struct {
  Idle      bool
  Destroyed bool
  LastUseAt time.Time
  UseCount  int
  Resource  interface{}
  Id        int64
}

type Config struct {
  Creator   CreatorFunc
  Destroyer DestroyerFunc
}

type Options struct {
  Min  int
  Max  int
  Idle int64
}

type ByTime []*Resource

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].LastUseAt.UnixNano() > a[j].LastUseAt.UnixNano() }

/**
Create a new pool
 */
func New(c Config, o Options) (p *Pool, err error) {

  if c.Creator == nil {
    err = errors.New("creator or poll must be an function")
    return
  }

  if c.Destroyer == nil {
    err = errors.New("destroyer or poll must be an function")
    return
  }

  if o.Min < 0 {
    o.Min = 0
  }

  if o.Max < 0 {
    o.Max = o.Min + 1
  }

  if o.Idle <= 0 {
    o.Idle = 30
  }

  if o.Max < o.Min {
    err = errors.New("the options of max must greater then or equal min")
    return
  }

  p = &Pool{
    Config:  c,
    Options: o,
  }
  ticker := time.NewTicker(time.Second)
  go func() {
    for _ = range ticker.C {
      p.checkIdle()
    }
  }()
  return
}

func (p *Pool) checkIdle() {
  lengthOfPool := len(p.Pool)
  if lengthOfPool > 0 && lengthOfPool > p.Options.Min {
    for i, resource := range p.Pool {
      idleAt := resource.LastUseAt.Unix() + p.Options.Idle
      now := time.Now().Unix()
      if now > idleAt {
        // over the min resource number, other should be destroy
        if i+1 > p.Options.Min {
          resource.Destroyed = true
        } else {
          // mark as idle resource
          resource.Idle = true
        }
      }
    }

    ret := make([]*Resource, 0)

    for _, resource := range p.Pool {
      if resource.Destroyed == false {
        ret = append(ret, resource)
      } else {
        if err := p.Config.Destroyer(p, resource.Resource); err != nil {
          // destroy fail
          resource.Destroyed = false
          ret = append(ret, resource)
        }
      }
    }

    sort.Sort(ByTime(ret))

    p.Pool = ret
  }
}

/**
Get entity
 */
func (p *Pool) Get() (interface{}, error) {
  if p.Destroyed == true {
    err := errors.New("the pool have been destroyed")
    return nil, err
  }

  if len(p.Pool) < p.Options.Max {
    id := p.index + 1
    if resource, err := p.Config.Creator(p, id); err != nil {
      return nil, err
    } else {
      p.index = id
      p.Pool = append(p.Pool, &Resource{
        LastUseAt: time.Now(),
        UseCount:  0,
        Resource:  resource,
        Idle:      false,
        Id:        p.index,
      })
      return resource, nil
    }
  } else {
    // if overload, them return the current resource
    sort.Sort(ByTime(p.Pool))
    p.Pool[0].LastUseAt = time.Now()
    p.Pool[0].Idle = false
    p.Pool[0].UseCount = p.Pool[0].UseCount + 1
    return *p.Pool[0], nil
  }
}

/**
Release a resource by id
 */
func (p *Pool) Release(id int64) (err error) {
  newPool := make([]*Resource, 0)

  defer func() {
    for _, resource := range p.Pool {
      if resource.Destroyed == false {
        newPool = append(newPool, resource)
      }
    }
    p.Pool = newPool
  }()

  for _, resource := range p.Pool {
    if resource.Id == id {
      if err = p.Config.Destroyer(p, resource.Resource); err == nil {
        resource.Destroyed = true
      }
    }
  }
  return
}

/**
Release the resource
 */
func (p *Pool) Destroy() (err error) {
  newPool := make([]*Resource, 0)

  defer func() {
    for _, resource := range p.Pool {
      if resource.Destroyed == false {
        newPool = append(newPool, resource)
      }
    }
    p.Pool = newPool
  }()

  for _, resource := range p.Pool {
    // destroy the resource
    if err = p.Config.Destroyer(p, resource.Resource); err != nil {
      resource.Destroyed = false
      return
    }
    resource.Destroyed = true
  }

  p.Destroyed = true
  return
}
