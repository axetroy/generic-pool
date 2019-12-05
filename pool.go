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
	Pool      *SafeMap
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
		Pool:    NewSafeMap(),
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
	if p.Pool.Count() > 0 && p.Pool.Count() > p.Options.Min {
		for id, resource := range p.Pool.Items() {
			r := resource.(*Resource)
			idleAt := r.LastUseAt.Unix() + p.Options.Idle
			now := time.Now().Unix()
			if now > idleAt {
				// over the min resource number, other should be destroy
				if p.Pool.Count() > p.Options.Min {
					if err := p.Config.Destroyer(p, r.Resource); err == nil {
						if ok := p.Pool.Check(id); ok {
							p.Pool.Delete(id)
						}
					}
				} else {
					// mark as idle resource
					r.Idle = true
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
	if p.Pool.Count() < p.Options.Max {
		id := p.index + 1
		if resource, err := p.Config.Creator(p, id); err != nil {
			return nil, err
		} else {
			p.index = id
			p.Pool.Set(id, &Resource{
				LastUseAt: time.Now(),
				UseCount:  0,
				Resource:  resource,
				Idle:      false,
				Id:        p.index,
			})
			return resource, nil
		}
	} else {
		// if overload, then get the latest resource
		var (
			lastUseAt       *time.Time
			lastUseResource *Resource
		)

		// find the resource which not use at latest
		for _, resource := range p.Pool.Items() {
			r := resource.(*Resource)
			if lastUseAt == nil {
				lastUseAt = &r.LastUseAt
				lastUseResource = r
			}
			if r.LastUseAt.UnixNano() < (*lastUseAt).UnixNano() {
				lastUseAt = &r.LastUseAt
				lastUseResource = r
				break
			}
		}

		if lastUseResource != nil {
			lastUseResource.LastUseAt = time.Now()
			lastUseResource.Idle = false
			lastUseResource.UseCount = lastUseResource.UseCount + 1
		}

		return lastUseResource, nil
	}
}

/**
Release a resource by id
*/
func (p *Pool) Release(id Id) (err error) {
	for resourceId, resource := range p.Pool.Items() {
		if resourceId == id {
			r := resource.(*Resource)
			if err = p.Config.Destroyer(p, r.Resource); err == nil {
				if ok := p.Pool.Check(id); ok {
					p.Pool.Delete(id)
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
	for id := range p.Pool.Items() {
		realId := id.(Id)
		if err = p.Release(realId); err != nil {
			return
		}
	}
	p.Destroyed = true
	return
}
