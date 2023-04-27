package game

import (
	"errors"
	"image"
	"log"
)

type DispatchFunc func(Action)

type Game struct {
	store    Store
	dispatch DispatchFunc
}

func NewGame(store Store, dispatch DispatchFunc) *Game {
	return &Game{
		store:    store,
		dispatch: dispatch,
	}
}

func (g *Game) HandleAction(action Action) {
	switch a := action.(type) {
	case PlayerInitSuccessAction:
		g.handlePlayerInitSuccessAction(a)
	case SpawnUnitAction:
		g.handleSpawnUnitAction(a)
	case MoveStartAction:
		g.handleMoveStartAction(a)
	case MoveStepAction:
		g.handleMoveStepAction(a)
	case MoveStopAction:
		g.handleMoveStopAction(a)
	case MapLoadSuccessAction:
		g.handleMapLoadSuccessAction(a)
	}
}

func (g *Game) handlePlayerInitSuccessAction(action PlayerInitSuccessAction) {
	for _, u := range action.Payload.Units {
		unit := u
		unit.dispatch = g.dispatch
		g.store.StoreUnit(unit)
		if err := g.placeUnit(unit); err != nil {
			log.Println(err)
		}
	}
	for _, p := range action.Payload.Players {
		player := p
		g.store.StorePlayer(player)
	}
}

func (g *Game) handleSpawnUnitAction(action SpawnUnitAction) {
	unit := action.Payload
	unit.dispatch = g.dispatch
	g.store.StoreUnit(unit)
	if err := g.placeUnit(unit); err != nil {
		log.Println(err)
		//dispatch error action
	}
}

func (g *Game) handleMoveStartAction(action MoveStartAction) {
	unit := g.store.GetUnitById(action.Payload.UnitId)

	unit.MoveTo(action.Payload.Point)
}

func (g *Game) handleMoveStepAction(action MoveStepAction) {
	//clean position
	for _, tile := range g.store.GetTilesByUnitId(action.Payload.UnitId) {
		tile.UnitId = ZeroUnitId
	}
	unit := g.store.GetUnitById(action.Payload.UnitId)

	unit.Position = action.Payload.Position
	unit.Path = action.Payload.Path
	unit.Step = action.Payload.Step

	if err := g.placeUnit(*unit); err != nil {
		log.Println(err)
		//dispatch error action
	}
	//reserve next step
	if len(action.Payload.Path) > action.Payload.Step {
		nextStep := action.Payload.Path[action.Payload.Step]
		if err := g.placeUnit(*unit, nextStep); err != nil {
			g.dispatch(MoveStopAction{
				Type:    MoveStopActionType,
				Payload: unit.Id,
			})
			/*
				Retry moving to target in 1s
				go func(a MoveUnitAction) {
					time.Sleep(1 * time.Second)
					a.Step -= 1
					if err := g.handleAction(a); err != nil {
						log.Println(err)
					}
				}(action) */
		}
	}
}

func (g *Game) handleMoveStopAction(action MoveStopAction) {
	unit := g.store.GetUnitById(action.Payload)

	unit.Path = []image.Point{}
	unit.Step = 0

}

func (g *Game) handleMapLoadSuccessAction(action MapLoadSuccessAction) {
	for _, t := range action.Payload.Tiles {
		g.store.StoreTile(t)
	}
}

func (g *Game) placeUnit(unit Unit, positions ...image.Point) error {
	if len(positions) == 0 {
		positions = []image.Point{unit.Position.ImagePoint()}
	}
	for _, p := range positions {
		t := g.store.GetTile(p)

		//set position
		if t.UnitId != ZeroUnitId && t.UnitId != unit.Id {
			return errors.New("position")
		}
		t.UnitId = unit.Id
	}

	return nil
}
