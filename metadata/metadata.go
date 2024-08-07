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

type (
	metaInKey  struct{}
	metaOutKey struct{}
)

// MD is a mapping from metadata keys to values.
// MD 是附加传输的上下文信息
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

// NewIncomingContext creates a new context with incoming md attached.
// Note: md must not be modified after calling this function.
// Note: md 不能在调用此函数后被修改
func NewIncomingContext(father context.Context, m map[string]any) context.Context {
	md := New(m)
	return context.WithValue(father, metaInKey{}, md)
}

// NewOutContext creates a new context with outgoing md attached.
// Note: md must not be modified after calling this function.
// Note: md 不能在调用此函数后被修改
func NewOutContext(father context.Context, m map[string]any) context.Context {
	md := New(m)
	return context.WithValue(father, metaOutKey{}, md)
}

func FromIncomingContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(metaInKey{}).(MD)
	if !ok {
		return nil, false
	}
	return
}

func FromOutContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(metaOutKey{}).(MD)
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
