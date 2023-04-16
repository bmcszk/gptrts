package main

import (
	"log"

	"github.com/bmcszk/gptrts/pkg/game"
	"github.com/bmcszk/gptrts/pkg/world"
)

type Game struct {
	*game.Game
	dispatch     game.DispatchFunc
	worldService world.WorldService
}

func NewGame(dispatch game.DispatchFunc, worldService world.WorldService) *Game {
	g := &Game{
		Game:         game.NewGame(dispatch),
		dispatch:     dispatch,
		worldService: worldService,
	}

	return g
}

func (g *Game) HandleAction(action game.Action) {
	log.Printf("server handle %s", action.GetType())
	g.Game.HandleAction(action)
	switch a := action.(type) {
	case game.PlayerInitAction:
		g.handlePlayerInitAction(a)
	case game.MapLoadAction:
		g.handleMapLoadAction(a)
	}
}

func (g *Game) handlePlayerInitAction(action game.PlayerInitAction) {
	player := &action.Payload
	g.Game.Players[action.Payload.Id] = player

	successAction := game.PlayerInitSuccessAction{
		Type: game.PlayerInitSuccessActionType,
		Payload: game.PlayerInitSuccessPayload{
			PlayerId: player.Id,
			Units:    make([]game.Unit, 0),
			Players:  make([]game.Player, 0),
		},
	}
	for _, unit := range g.Game.Units {
		successAction.Payload.Units = append(successAction.Payload.Units, *unit)
	}
	for _, player := range g.Game.Players {
		successAction.Payload.Players = append(successAction.Payload.Players, *player)
	}
	g.dispatch(successAction)

	var startingP game.PF
	for sp, p := range g.Starting {
		if p == nil {
			startingP = sp
			g.Starting[sp] = &player.Id
			break
		}
	}
	unit := game.NewUnit(action.Payload.Id, player.Color, startingP, 32, 32)
	g.Units[unit.Id] = unit // should it driven by action?
	unitAction := game.AddUnitAction{
		Type:    game.AddUnitActionType,
		Payload: *unit,
	}
	g.dispatch(unitAction)
}

func (g *Game) handleMapLoadAction(action game.MapLoadAction) {
	resp, err := g.worldService.Load(action.Payload.WorldRequest)
	if err != nil {
		log.Printf("error loading map: %s", err)
		return
		// TODO: error handling
		// TODO: send error to client
	}
	successAction := game.MapLoadSuccessAction{
		Type: game.MapLoadSuccessActionType,
		Payload: game.MapLoadSuccessPayload{
			WorldResponse: *resp,
			PlayerId:      action.Payload.PlayerId,
		},
	}
	g.dispatch(successAction)
}
