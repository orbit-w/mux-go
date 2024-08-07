package metadata

import (
	"context"
	"testing"
)

/*
   @Author: orbit-w
   @File: metadata_test
   @2024 8月 周三 21:35
*/

func Test_FromMetaContext(t *testing.T) {
	md := New(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	ctx := NewMetaContext(context.Background(), md)
	md2, ok := FromMetaContext(ctx)
	if !ok {
		t.Fatal("metadata not exist")
	}
	t.Log(md2)
}

func TestMD_Get(t *testing.T) {
	md := New(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	v, exist := md.Get("key1")
	if !exist {
		t.Fatal("key1 not exist")
	}
	if v != "value1" {
		t.Fatal("key1 value not equal")
	}
}

func TestMD_Set(t *testing.T) {
	md := New(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	md.Set("key1", "value3")
	v, exist := md.Get("key1")
	if !exist {
		t.Fatal("key1 not exist")
	}
	if v != "value3" {
		t.Fatal("key1 value not equal")
	}
}

func Test_Serialize(t *testing.T) {
	md := New(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	data, err := Marshal(md)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))

	var md2 MD
	err = Unmarshal(data, &md2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(md2)
}
