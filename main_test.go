package main

import (
	"testing"

	"gotest.tools/assert"
)

func TestGenid(t *testing.T) {
	set := map[string]bool{}
	cnt := 100000
	for i := 0; i < cnt; i++ {
		id := genid()
		set[id] = true
	}
	assert.Equal(t, len(set), cnt)
	i := 0
	for k, v := range set {
		i++
		t.Log(k, v)
		if i > 100 {
			return
		}
	}
}
