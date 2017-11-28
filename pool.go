package pool

import (
  "errors"
  "time"
  "sort"
  "fmt"
)

type CreatorFunc func(p *Pool) (resource interface{}, err error)
type DestroyerFunc func(p *Pool, resource interface{}) (err error)

type Pool struct {
  config    Config
  options   Options
  pool      []*Resource
  destroyed bool
}

type Resource struct {
  Idle      bool
  Destroyed bool
  LastUseAt time.Time
  UseCount  int
  Resource  interface{}
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
func New(c Config, o Options) (p *Pool) {
  p = &Pool{
    config:  c,
    options: o,
  }
  ticker := time.NewTicker(time.Second * 1)
  go func() {
    for _ = range ticker.C {
      p.checkIdle()
    }
  }()
  return
}

/**
Remove Resource from pool
 */
func (p *Pool) RemoveFromPool(i int) {
  if len(p.pool) == 1 {
    p.pool = make([]*Resource, 0)
  } else {
    p.pool = append(p.pool[:i], p.pool[:i+1]...)
  }
}

func (p *Pool) checkIdle() {
  lengthOfPool := len(p.pool)
  if lengthOfPool > 0 && lengthOfPool > p.options.Min {
    for i, resource := range p.pool {
      idleAt := resource.LastUseAt.Unix() + p.options.Idle
      now := time.Now().Unix()
      if now > idleAt {
        // over the min resource number, other should be destroy
        if i+1 > p.options.Min {
          resource.Destroyed = true
        } else {
          // mark as idle resource
          resource.Idle = true
        }
      }
    }

    ret := make([]*Resource, 0)

    for _, resource := range p.pool {
      if resource.Destroyed == false {
        ret = append(ret, resource)
      } else {
        if err := p.config.Destroyer(p, resource.Resource); err != nil {
          // destroy fail
          fmt.Println(err)
        }
      }
    }

    sort.Sort(ByTime(ret))

    p.pool = ret
  }
}

/**
Get entity
 */
func (p *Pool) Get() (interface{}, error) {
  if p.destroyed == true {
    err := errors.New("the pool have been destroyed")
    return nil, err
  }

  if len(p.pool) < p.options.Max {
    if resource, err := p.config.Creator(p); err != nil {
      return nil, err
    } else {
      p.pool = append(p.pool, &Resource{
        LastUseAt: time.Now(),
        UseCount:  0,
        Resource:  resource,
        Idle:      false,
      })
      return resource, nil
    }
  } else {
    // if overload, them return the current resource
    sort.Sort(ByTime(p.pool))
    p.pool[0].LastUseAt = time.Now()
    p.pool[0].Idle = false
    return *p.pool[0], nil
  }
}

/**
Release the resource
 */
func (p *Pool) Destroy() (err error) {
  for i, resource := range p.pool {
    // destroy the resource
    if err = p.config.Destroyer(p, resource.Resource); err != nil {
      return
    }
    // remove from pool

    // if length=1, set it to empty array
    if len(p.pool) == 1 {
      p.pool = make([]*Resource, 0)
    } else {
      p.pool = append(p.pool[:i], p.pool[:i+1]...)
    }
  }
  p.destroyed = true
  return
}
