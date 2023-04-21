package main

import (
	"image"
	"image/color"
	"log"
	"sync"

	"github.com/bmcszk/gptrts/pkg/convert"
	"github.com/bmcszk/gptrts/pkg/game"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	cameraSpeed = 2
)

type Game struct {
	*game.Game
	Map              *Map
	Units            map[game.UnitIdType]*Unit
	PlayerId         game.PlayerIdType
	cameraX, cameraY int
	centerX, centerY int
	selectionBox     *image.Rectangle
	dispatch         game.DispatchFunc
	mux              *sync.Mutex
	screen           *Screen
}

func NewGame(playerId game.PlayerIdType, dispatch game.DispatchFunc) *Game {
	g := game.NewGame(dispatch)
	cg := &Game{
		PlayerId: playerId,
		Game:     g,
		Map:      NewMap(g.Map),
		Units:    make(map[game.UnitIdType]*Unit),
		dispatch: dispatch,
		mux:      &sync.Mutex{},
		screen:   NewScreen(),
	}

	return cg
}

func (g *Game) SetMap(m *game.Map) {
	g.Game.SetMap(m)
	g.Map = NewMap(m)
}

func (g *Game) SetUnit(unit *game.Unit) {
	g.Game.SetUnit(unit)
	g.Units[unit.Id] = NewUnit(unit)
	g.Map.UpdateVisibility(unit)
	for _, u := range g.Units {
		u.UpdateVisibility(unit)
	}
}

func (g *Game) SetPlayer(player *game.Player) {
	g.Game.SetPlayer(player)
}

func (g *Game) HandleAction(action game.Action) {
	g.mux.Lock()
	defer g.mux.Unlock()

	log.Printf("client handle %s", action.GetType())
	g.Game.HandleAction(action)
	switch a := action.(type) {
	case game.AddUnitAction:
		g.handleAddUnitAction(a)
	case game.PlayerInitSuccessAction:
		g.handlePlayerInitSuccessAction(a)
	case game.MapLoadSuccessAction:
		g.handleMapLoadSuccessAction(a)
	case game.MoveStepAction:
		g.handleMoveStepAction(a)
	}
}

func (g *Game) handleAddUnitAction(action game.AddUnitAction) {
	g.SetUnit(&action.Payload)
}

func (g *Game) handlePlayerInitSuccessAction(action game.PlayerInitSuccessAction) {
	for _, u := range action.Payload.Units {
		unit := u
		g.SetUnit(&unit)
	}
	for _, p := range action.Payload.Players {
		player := p
		g.SetPlayer(&player)
	}
}

func (g *Game) handleMapLoadSuccessAction(action game.MapLoadSuccessAction) {
	for _, t := range action.Payload.Tiles {
		tile := t
		g.Map.SetTile(&tile)
	}
	g.loadMap()
}

func (g *Game) handleMoveStepAction(action game.MoveStepAction) {
	unit := g.Game.Units[action.Payload.UnitId]
	g.Map.UpdateVisibility(unit)
	for _, u := range g.Units {
		u.UpdateVisibility(unit)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	/* // Calculate the desired screen size based on the size of the map
	sw := len(g.Map.Tiles[0]) * tileSize
	sh := len(g.Map.Tiles) * tileSize

	// Scale the screen if it is too large to fit
	if sw > outsideWidth || sh > outsideHeight {
		scale := math.Min(float64(outsideWidth)/float64(sw), float64(outsideHeight)/float64(sh))
		sw = int(float64(sw) * scale)
		sh = int(float64(sh) * scale)
	}

	return sw, sh */

	// g.centerX = -outsideWidth /2
	// g.centerY = -outsideHeight /2

	minX, minY := g.screenToWorldTiles(0, 0)
	maxX, maxY := g.screenToWorldTiles(outsideWidth, outsideHeight)
	rect := image.Rect(minX, minY, maxX, maxY)

	if g.SetScreen(rect) {
		g.loadMap()
	}

	return outsideWidth, outsideHeight
}

func (g *Game) SetScreen(rect image.Rectangle) bool {
	if g.screen.rect.Eq(rect) {
		return false
	}
	currRect := g.screen.rect
	if rect.Min.X < currRect.Min.X {
		action := game.NewMapLoadAction(image.Rect(rect.Min.X, rect.Min.Y, currRect.Min.X, currRect.Max.Y), g.PlayerId)
		g.dispatch(action)
	}
	if rect.Min.Y < currRect.Min.Y {
		action := game.NewMapLoadAction(image.Rect(rect.Min.X, rect.Min.Y, currRect.Max.X, currRect.Min.Y), g.PlayerId)
		g.dispatch(action)
	}
	if rect.Max.X > currRect.Max.X {
		action := game.NewMapLoadAction(image.Rect(currRect.Min.X, currRect.Max.Y, rect.Max.X, rect.Max.Y), g.PlayerId)
		g.dispatch(action)
	}
	if rect.Max.Y > currRect.Max.Y {
		action := game.NewMapLoadAction(image.Rect(currRect.Max.Y, currRect.Min.Y, rect.Max.X, rect.Max.Y), g.PlayerId)
		g.dispatch(action)
	}
	g.screen.rect = rect
	return true
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the map
	g.screen.Draw(screen, g.centerX+g.cameraX, g.centerY+g.cameraY)

	// Draw units
	for _, unit := range g.Units {
		unit.Draw(screen, g.centerX+g.cameraX, g.centerY+g.cameraY)
	}

	// Draw the selection box
	if g.selectionBox != nil {
		r := *g.selectionBox
		x1, y1 := g.worldToScreen(r.Min.X, r.Min.Y)
		x2, y2 := g.worldToScreen(r.Max.X, r.Max.Y)

		col := color.RGBA{0, 255, 0, 128}
		ebitenutil.DrawRect(screen, float64(x1), float64(y1), float64(x2-x1), float64(y2-y1), col)
	}
}

func (g *Game) Update() error {
	// Move camera with arrow keys
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.cameraX -= cameraSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.cameraX += cameraSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.cameraY -= cameraSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.cameraY += cameraSpeed
	}

	// Handle left mouse button click to select units
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && ebiten.IsFocused() {
		mx, my := ebiten.CursorPosition()
		worldX, worldY := g.screenToWorld(mx, my)

		if g.selectionBox == nil {
			g.selectionBox = convert.ToPointer(image.Rect(worldX, worldY, worldX+1, worldY+1))
		} else {
			g.selectionBox.Max = image.Pt(worldX+1, worldY+1)
		}
	} else {
		g.selectionBox = nil
	}

	if g.selectionBox != nil {
		r := *g.selectionBox
		for _, u := range g.Units {
			if r.Canon().Overlaps(u.GetRect()) {
				u.Selected = true
			} else if !ebiten.IsKeyPressed(ebiten.KeyShift) {
				u.Selected = false
			}
		}
	}

	// Handle right mouse button click to move selected units
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) && ebiten.IsFocused() {
		mx, my := ebiten.CursorPosition()
		tileX, tileY := g.screenToWorldTiles(mx, my)
		for _, u := range g.Units {
			if u.Selected && u.Unit.Owner == g.PlayerId {
				moveStartAction := game.MoveStartAction{
					Type: game.MoveStartActionType,
					Payload: game.MoveStartPayload{
						UnitId: u.Id,
						Point:  game.NewPF(float64(tileX), float64(tileY)),
					},
				}
				g.dispatch(moveStartAction)
			}
		}
	}

	for _, u := range g.Units {
		u.Update()
	}

	return nil
}

func (g *Game) screenToWorld(screenX, screenY int) (int, int) {
	worldX := screenX + g.cameraX
	worldY := screenY + g.cameraY
	return worldX, worldY
}

func (g *Game) screenToWorldTiles(screenX, screenY int) (int, int) {
	tileX := (screenX + g.cameraX) / tileSize
	tileY := (screenY + g.cameraY) / tileSize
	return tileX, tileY
}

func (g *Game) worldToScreen(worldX, worldY int) (int, int) {
	screenX := worldX - g.cameraX
	screenY := worldY - g.cameraY
	return screenX, screenY
}

func (g *Game) loadMap() {
	g.screen.tiles = make([][]*Tile, 0, 100 /*optimize me*/)
	for x := g.screen.rect.Min.X; x <= g.screen.rect.Max.X; x++ {
		row := make([]*Tile, 0, 100 /*optimize me*/)
		for y := g.screen.rect.Min.Y; y <= g.screen.rect.Max.Y; y++ {
			row = append(row, g.Map.GetTile(image.Pt(x, y)))
		}
		g.screen.tiles = append(g.screen.tiles, row)
	}
}
