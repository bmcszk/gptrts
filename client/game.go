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

type QueueFunc func(game.Action)

type Game struct {
	*game.Game
	store            game.Store
	PlayerId         game.PlayerIdType
	cameraX, cameraY int
	centerX, centerY int
	selectionBox     *image.Rectangle
	queueFunc        QueueFunc
	mux              *sync.Mutex
	screen           *Screen
}

func NewGame(playerId game.PlayerIdType, store game.Store, queueFunc QueueFunc) *Game {
	g := game.NewGame(store)
	cg := &Game{
		store:     store,
		PlayerId:  playerId,
		Game:      g,
		queueFunc: queueFunc,
		mux:       &sync.Mutex{},
		screen:    NewScreen(image.Rectangle{}),
	}

	return cg
}

func (g *Game) HandleAction(action game.Action, dispatch game.DispatchFunc) {
	log.Printf("client handle %s", action.GetType())
	g.Game.HandleAction(action, dispatch)
	switch action.(type) {
	case game.SpawnUnitAction, game.MoveStepAction, game.PlayerJoinSuccessAction:
		g.updateVisibility()
	case game.MapLoadSuccessAction:
		g.loadScreen()
		g.updateVisibility()
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
		g.loadScreen()
		g.updateVisibility()
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
		g.queueFunc(action)
	}
	if rect.Min.Y < currRect.Min.Y {
		action := game.NewMapLoadAction(image.Rect(rect.Min.X, rect.Min.Y, currRect.Max.X, currRect.Min.Y), g.PlayerId)
		g.queueFunc(action)
	}
	if rect.Max.X > currRect.Max.X {
		action := game.NewMapLoadAction(image.Rect(currRect.Min.X, currRect.Max.Y, rect.Max.X, rect.Max.Y), g.PlayerId)
		g.queueFunc(action)
	}
	if rect.Max.Y > currRect.Max.Y {
		action := game.NewMapLoadAction(image.Rect(currRect.Max.Y, currRect.Min.Y, rect.Max.X, rect.Max.Y), g.PlayerId)
		g.queueFunc(action)
	}
	g.screen.rect = rect
	return true
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw the map
	g.screen.Draw(screen, g.centerX+g.cameraX, g.centerY+g.cameraY)

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
		for _, u := range g.store.GetAllUnits() {
			if r.Canon().Overlaps(getRect(u)) {
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
		for _, u := range g.store.GetAllUnits() {
			if u.Selected && u.Owner == g.PlayerId {
				moveStartAction := game.MoveStartAction{
					Type: game.MoveStartActionType,
					Payload: game.MoveStartPayload{
						UnitId: u.Id,
						Point:  image.Pt(tileX, tileY),
					},
				}
				g.queueFunc(moveStartAction)
			}
		}
	}

	for _, u := range g.store.GetAllUnits() {
		u.Update()
	}

	return nil
}

func (g *Game) updateVisibility() {
	m := make(map[image.Point]bool, 0)
	for _, t := range g.screen.tiles {
		if t.Visible || t.Unit != nil {
			m[t.Point] = false
		}
	}
	for _, unit := range g.store.GetAllUnits() {
		if unit.Owner != g.PlayerId {
			continue
		}
		for _, vector := range unit.ISee {
			p := unit.Position.ImagePoint().Add(vector)
			m[p] = true
		}
	}
	for p, v := range m {
		if t, ok := g.screen.tiles[p]; ok {
			t.Visible = v
		}
	}
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

func (g *Game) loadScreen() {
	screen := NewScreen(g.screen.rect)
	for x := screen.rect.Min.X; x < screen.rect.Max.X; x++ {
		for y := screen.rect.Min.Y; y < screen.rect.Max.Y; y++ {
			p := image.Pt(x, y)
			t, ok := g.store.GetTile(p)
			if !ok {
				t = g.store.CreateTile(p)
			}
			screen.tiles[p] = t
		}
	}
	g.screen = screen
}
