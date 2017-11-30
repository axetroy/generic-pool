package pool

import "testing"

var safeMap *SafeMap

func TestNewSafeMap(t *testing.T) {
  safeMap = NewSafeMap()
  if safeMap == nil {
    t.Fatal("expected to return non-nil SafeMap", "got", safeMap)
  }
}

func TestSet(t *testing.T) {
  if ok := safeMap.Set("astaxie", 1); !ok {
    t.Error("expected", true, "got", false)
  }
}

func TestReSet(t *testing.T) {
  safeMap := NewSafeMap()
  if ok := safeMap.Set("aabbcc", 1); !ok {
    t.Error("expected", true, "got", false)
  }
  // set diff value
  if ok := safeMap.Set("aabbcc", -1); !ok {
    t.Error("expected", true, "got", false)
  }

  // set same value
  if ok := safeMap.Set("aabbcc", -1); ok {
    t.Error("expected", false, "got", true)
  }
}

func TestCheck(t *testing.T) {
  if exists := safeMap.Check("astaxie"); !exists {
    t.Error("expected", true, "got", false)
  }
}

func TestGet(t *testing.T) {
  if val := safeMap.Get("astaxie"); val.(int) != 1 {
    t.Error("expected value", 1, "got", val)
  }

  if val := safeMap.Get("not exist key"); val != nil {
    t.Error("Should get an value nil")
  }
}

func TestDelete(t *testing.T) {
  safeMap.Delete("astaxie")
  if exists := safeMap.Check("astaxie"); exists {
    t.Error("expected element to be deleted")
  }
}

func TestItems(t *testing.T) {
  safeMap := NewSafeMap()
  safeMap.Set("hello", "world")
  for k, v := range safeMap.Items() {
    key := k.(string)
    value := v.(string)
    if key != "hello" {
      t.Error("expected the key should be hello")
    }
    if value != "world" {
      t.Error("expected the value should be world")
    }
  }
}

func TestCount(t *testing.T) {
  if count := safeMap.Count(); count != 0 {
    t.Error("expected count to be", 0, "got", count)
  }
}
