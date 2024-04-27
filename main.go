package main

import (
	"errors"
	"strings"
	"sync"

// 	"golang.org/x/text/cases"
// 	"golang.org/x/text/language"
)

type Player struct {
	name string
	m    *Map
	ch   chan string
}

type Map struct {
	id      int
	players []Player
	ch      chan string
}

type Game struct {
	maps    []Map
	players []Player
}

func CapitalizeFirstLetter(word string) string {
	if word == "" {
		return ""
	}
	return strings.ToUpper(word[:1]) + word[1:]
}

func (m *Map) deletePlayerFromMap(name string) {
	var mu sync.Mutex
	for i, player := range m.players {
		if player.name == strings.ToLower(name) {
			mu.Lock()
			m.players = append(m.players[:i], m.players[i+1:]...)
			mu.Unlock()
		}
	}
}

func (g *Game) returnMap(mapId int) *Map {
	for _, m := range g.maps {
		if m.id == mapId {
			return &m
		}
	}
	return nil
}

func NewGame(mapIds []int) (*Game, error) {
	g := Game{
		maps:    make([]Map, 0),
		players: make([]Player, 0),
	}

	for _, mapId := range mapIds {
		if mapId <= 0 {
			return nil, errors.New("Map id cannot be negative or zero")
		}
		var m Map = Map{
			id:      mapId,
			players: make([]Player, 0),
			ch:      make(chan string),
		}

		g.maps = append(g.maps, m)
	}
	return &g, nil
}

func (g *Game) ConnectPlayer(name string) error {
	var mu sync.Mutex
	lowerName := strings.ToLower(name)
	for _, player := range (*g).players {
		if player.name == lowerName {
			return errors.New("Player already exists")
		}
	}
	p := Player{
		name: lowerName,
		m:    &((*g).maps[0]),
		ch:   make(chan string),
	}
	mu.Lock()
	(*g).maps[0].players = append((*g).maps[0].players, p)
	(*g).players = append((*g).players, p)
	mu.Unlock()
	return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
	var mu sync.Mutex
	idFlag := false
	for _, m := range (*g).maps {
		if m.id == mapId {
			idFlag = true
		}
	}
	m := g.returnMap(mapId)
	if !idFlag {
		return errors.New("Map not found")
	}
	for _, p := range (*g).players {
		if p.name == strings.ToLower(name) {
			mu.Lock()
			(p.m).deletePlayerFromMap(p.name)
			p.m = m
			mu.Unlock()
			return nil
		}
	}
	return errors.New("Player not found")
}

func (g *Game) GetPlayer(name string) (*Player, error) {
	lowerName := strings.ToLower(name)
	for _, p := range (*g).players {
		if p.name == lowerName {
			return &p, nil
		}
	}
	return nil, errors.New("Player not found")
}

func (g *Game) GetMap(mapId int) (*Map, error) {
	for _, m := range (*g).maps {
		if m.id == mapId {
			return &m, nil
		}
	}
	return nil, errors.New("Map not found")
}

func (m *Map) FanOutMessages() {
    var mu sync.Mutex
	select {
	case msg := <-m.ch:
		name, _, _ := strings.Cut(msg, " says:")
		for _, p := range m.players {
			if p.name != strings.ToLower(name) {
			    mu.Lock()
				p.ch <- msg
				mu.Unlock()
			}
		}
	}
}

func (p *Player) GetChannel() <-chan string {
	return (*p).ch
}

func (p *Player) SendMessage(msg string) error {
    var mu sync.Mutex
    if len(msg) == 0 {
        return errors.New("Your msg is empty")
    }
	msg = CapitalizeFirstLetter(p.name) + " says: " + msg
	mu.Lock()
	p.m.ch <- msg
	mu.Unlock()
	return nil
}

func (p *Player) GetName() string {
	return (*p).name
}
