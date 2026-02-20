package storage

import (
	"context"
	"crypto/rand"
	"sync"
	"time"
)

type tempToken = [32]byte

type credentials struct {
	email    string
	password string
}

type tempEntry struct {
	cred      credentials
	expiresAt time.Time
}

type gateway struct {
	mx sync.RWMutex
	mp map[tempToken]tempEntry
}

func generateTempToken() tempToken {
	var t tempToken
	if _, err := rand.Read(t[:]); err != nil {
		panic(err)
	}
	return t
}

func newGateway(ctx context.Context) *gateway {
	g := &gateway{
		mx: sync.RWMutex{},
		mp: map[tempToken]tempEntry{},
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
	toDelete := make([]tempToken, 0, len(g.mp))
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

func (g *gateway) Store(c credentials) tempToken {
	t := generateTempToken()
	g.mx.Lock()
	g.mp[t] = tempEntry{
		cred:      c,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	g.mx.Unlock()
	return t
}

func (g *gateway) Get(token tempToken) (credentials, bool) {
	g.mx.RLock()
	entry, ok := g.mp[token]
	g.mx.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			g.mx.Lock()
			delete(g.mp, token)
			g.mx.Unlock()
		}
		return credentials{}, false
	}

	return entry.cred, true
}

func (g *gateway) Erase(token tempToken, cred credentials) {
	g.mx.Lock()
	entry, ok := g.mp[token]
	if ok && entry.cred == cred {
		delete(g.mp, token)
	}
	g.mx.Unlock()
}
