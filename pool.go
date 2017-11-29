package pool

import (
  "errors"
  "time"
)

type Id int64

type CreatorFunc func(p *Pool, id Id) (resource interface{}, err error)
type DestroyerFunc func(p *Pool, resource interface{}) (err error)

type Pool struct {
  Config    Config
  Options   Options
  Pool      map[Id]*Resource
  Destroyed bool
  index     Id
}

type Resource struct {
  Idle      bool
  LastUseAt time.Time
  UseCount  int
  Resource  interface{}
  Id        Id
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
    Pool:    map[Id]*Resource{},
  }
  ticker := time.NewTicker(time.Second)
  go func() {
    for _ = range ticker.C {
      p.checkIdle()
    }
  }()
  return
}

/**
Check the pool idle resource
 */
func (p *Pool) checkIdle() {
  lengthOfPool := len(p.Pool)
  if lengthOfPool > 0 && lengthOfPool > p.Options.Min {
    for id, resource := range p.Pool {
      idleAt := resource.LastUseAt.Unix() + p.Options.Idle
      now := time.Now().Unix()
      if now > idleAt {
        // over the min resource number, other should be destroy
        if len(p.Pool) > p.Options.Min {
          if err := p.Config.Destroyer(p, resource.Resource); err == nil {
            if _, ok := p.Pool[id]; ok {
              delete(p.Pool, id)
            }
          }
        } else {
          // mark as idle resource
          resource.Idle = true
        }
      }
    }
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

  // the pool not fulled, it should create new one
  if len(p.Pool) < p.Options.Max {
    id := p.index + 1
    if resource, err := p.Config.Creator(p, id); err != nil {
      return nil, err
    } else {
      p.index = id
      p.Pool[id] = &Resource{
        LastUseAt: time.Now(),
        UseCount:  0,
        Resource:  resource,
        Idle:      false,
        Id:        p.index,
      }
      return resource, nil
    }
  } else {
    // if overload, then get the latest resource
    var (
      lastUseAt       *time.Time
      lastUseResource *Resource
    )

    // find the resource which not use at latest
    for _, resource := range p.Pool {
      if lastUseAt == nil {
        lastUseAt = &resource.LastUseAt
        lastUseResource = resource
      }
      if resource.LastUseAt.UnixNano() < (*lastUseAt).UnixNano() {
        lastUseAt = &resource.LastUseAt
        lastUseResource = resource
        break
      }
    }
    lastUseResource.LastUseAt = time.Now()
    lastUseResource.Idle = false
    lastUseResource.UseCount = lastUseResource.UseCount + 1

    return lastUseResource, nil
  }
}

/**
Release a resource by id
 */
func (p *Pool) Release(id Id) (err error) {
  for resourceId, resource := range p.Pool {
    if resourceId == id {
      if err = p.Config.Destroyer(p, resource.Resource); err == nil {
        if _, ok := p.Pool[id]; ok {
          delete(p.Pool, id)
        }
      }
    }
  }
  return
}

/**
Release the resource
 */
func (p *Pool) Destroy() (err error) {
  for id := range p.Pool {
    if err = p.Release(id); err != nil {
      return
    }
  }
  p.Destroyed = true
  return
}
