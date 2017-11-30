// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pool

import (
  "sync"
)

// SafeMap is a map with lock
type SafeMap struct {
  lock *sync.RWMutex
  bm   map[interface{}]interface{}
}

// New return new safemap
func NewSafeMap() *SafeMap {
  return &SafeMap{
    lock: new(sync.RWMutex),
    bm:   make(map[interface{}]interface{}),
  }
}

// Get from maps return the k's value
func (m *SafeMap) Get(k interface{}) interface{} {
  m.lock.RLock()
  defer m.lock.RUnlock()
  if val, ok := m.bm[k]; ok {
    return val
  }
  return nil
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *SafeMap) Set(k interface{}, v interface{}) bool {
  m.lock.Lock()
  defer m.lock.Unlock()
  if val, ok := m.bm[k]; !ok {
    m.bm[k] = v
  } else if val != v {
    m.bm[k] = v
  } else {
    return false
  }
  return true
}

// Check Returns true if k is exist in the map.
func (m *SafeMap) Check(k interface{}) bool {
  m.lock.RLock()
  defer m.lock.RUnlock()
  _, ok := m.bm[k]
  return ok
}

// Delete the given key and value.
func (m *SafeMap) Delete(k interface{}) {
  m.lock.Lock()
  defer m.lock.Unlock()
  delete(m.bm, k)
}

// Items returns all items in safemap.
func (m *SafeMap) Items() map[interface{}]interface{} {
  m.lock.RLock()
  defer m.lock.RUnlock()
  r := make(map[interface{}]interface{})
  for k, v := range m.bm {
    r[k] = v
  }
  return r
}

// Count returns the number of items within the map.
func (m *SafeMap) Count() int {
  m.lock.RLock()
  defer m.lock.RUnlock()
  return len(m.bm)
}
