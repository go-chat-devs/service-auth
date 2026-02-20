package storage

import (
	"context"
	"sync"
	"time"

	"github.com/go-chat-devs/service-auth/internal/token"
)

type tempEntry struct {
	userId    int
	expiresAt time.Time
}

type gateway struct {
	mx sync.RWMutex
	mp map[token.Token]tempEntry
}

func newGateway(ctx context.Context) *gateway {
	g := &gateway{
		mx: sync.RWMutex{},
		mp: map[token.Token]tempEntry{},
	}
	go g.cleanupWorker(ctx)
	return g
}

func (g *gateway) cleanupWorker(ctx context.Context) {
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

func (g *gateway) cleanup() {
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

func (g *gateway) Store(userId int) token.Token {
	t := token.Generate()
	g.mx.Lock()
	g.mp[t] = tempEntry{
		userId:    userId,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	g.mx.Unlock()
	return t
}

func (g *gateway) Get(token token.Token) (int, bool) {
	g.mx.RLock()
	entry, ok := g.mp[token]
	g.mx.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			g.mx.Lock()
			delete(g.mp, token)
			g.mx.Unlock()
		}
		return 0, false
	}

	return entry.userId, true
}

func (g *gateway) Erase(token token.Token, userId int) {
	g.mx.Lock()
	entry, ok := g.mp[token]
	if ok && entry.userId == userId {
		delete(g.mp, token)
	}
	g.mx.Unlock()
}
