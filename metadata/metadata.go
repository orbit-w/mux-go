package metadata

import (
	"context"
	"encoding/json"
)

/*
   @Author: orbit-w
   @File: metadata
   @2023 11月 周日 17:15
*/

type metaDataKey struct{}

type MD map[string]any

func New(kvs map[string]any) MD {
	md := MD{}
	for k, v := range kvs {
		md[k] = v
	}
	return md
}

func (ins *MD) Set(key string, value any) {
	md := *ins
	md[key] = value
}

func (ins *MD) Get(key string) (v any, exist bool) {
	md := *ins
	v, exist = md[key]
	return
}

func NewMetaContext(father context.Context, m map[string]any) context.Context {
	md := New(m)
	return context.WithValue(father, metaDataKey{}, md)
}

func GetValue(ctx context.Context, key string) (v any, exist bool) {
	md, ok := FromMetaContext(ctx)
	if !ok {
		return nil, false
	}
	return md.Get(key)
}

func FromMetaContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(metaDataKey{}).(MD)
	if !ok {
		return nil, false
	}
	return
}

func Marshal(m MD) ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Unmarshal(data []byte, dst *MD) error {
	return json.Unmarshal(data, dst)
}
