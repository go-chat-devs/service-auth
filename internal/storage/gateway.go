package storage

import (
	"context"
	"sync"
	"time"

	"github.com/go-chat-devs/service-auth/internal/token"
)

type gateway[T any] struct {
	mx sync.RWMutex
	mp map[token.Token]gatewayEntry[T]
}

type gatewayEntry[T any] struct {
	val       T
	expiresAt time.Time
}

func newGateway[T any](ctx context.Context) *gateway[T] {
	g := &gateway[T]{
		mx: sync.RWMutex{},
		mp: map[token.Token]gatewayEntry[T]{},
	}
	go g.cleanupWorker(ctx)
	return g
}

func (g *gateway[T]) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.cleanup()
		}
	}
}

func (g *gateway[T]) cleanup() {
	t := time.Now()
	toDelete := make([]token.Token, 0, len(g.mp))
	g.mx.Lock()
	for k, v := range g.mp {
		if t.After(v.expiresAt) {
			toDelete = append(toDelete, k)
		}
	}
	for _, v := range toDelete {
		delete(g.mp, v)
	}
	g.mx.Unlock()
}

func (g *gateway[T]) Store(val T) token.Token {
	t := token.Generate()
	g.mx.Lock()
	g.mp[t] = gatewayEntry[T]{
		val:       val,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	g.mx.Unlock()
	return t
}

func (g *gateway[T]) Get(token token.Token) (val T, ok bool) {
	g.mx.RLock()
	entry, ok := g.mp[token]
	g.mx.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			g.mx.Lock()
			delete(g.mp, token)
			g.mx.Unlock()
		}
		ok = false
		return
	}
	ok = true
	val = entry.val
	return
}

func (g *gateway[T]) Erase(token token.Token) {
	g.mx.Lock()
	if _, ok := g.mp[token]; ok {
		delete(g.mp, token)
	}
	g.mx.Unlock()
}
