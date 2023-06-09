package main

import (
	"encoding/hex"
	"image"
	"image/color"
	"log"

	"github.com/bmcszk/gptrts/pkg/game"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	tileSize       = 16
	tileSpriteSize = 16
	tileSpriteXNum = 7
	selectedBorder = 2
)

type screen struct {
	rect  image.Rectangle
	tiles map[image.Point]*game.Tile
	units map[*game.Unit]bool
}

var emptyScreen = screen{
	rect:  image.Rectangle{},
	tiles: make(map[image.Point]*game.Tile),
	units: make(map[*game.Unit]bool),
}

func newScreen(rect image.Rectangle, tiles map[image.Point]*game.Tile) *screen {
	return &screen{
		rect:  rect,
		tiles: tiles,
		units: make(map[*game.Unit]bool, 0),
	}
}

func (s *screen) is(rect image.Rectangle) bool {
	return s.rect.Eq(rect)
}

func (s *screen) draw(enScreen *ebiten.Image, cameraX, cameraY int) {
	for _, t := range s.tiles {
		if t != nil {
			drawTile(t, enScreen, cameraX, cameraY)
			if t.Unit != nil {
				s.units[t.Unit] = t.Visible
			}
		}
	}
	for u, visible := range s.units {
		if visible {
			drawUnit(u, enScreen, cameraX, cameraY)
		}
	}
}

func drawTile(t *game.Tile, enScreen *ebiten.Image, cameraX, cameraY int) {
	p := t.Point

	op := &ebiten.DrawImageOptions{}

	if !t.Visible {
		// Create a new color matrix and set the brightness to a lower value
		cm := ebiten.ColorM{}
		cm.Scale(0.5, 0.5, 0.5, 1.0) // Make the tile darker
		op.ColorM = cm
	}

	op.GeoM.Translate(float64(p.X*tileSize-cameraX), float64(p.Y*tileSize-cameraY))
	enScreen.DrawImage(getBackgroundColorImage(t.BackStyleClass), op)

	tileSpriteNo := getTile(t.FrontStyleClass)
	sx := (tileSpriteNo % tileSpriteXNum) * tileSpriteSize
	sy := (tileSpriteNo / tileSpriteXNum) * tileSpriteSize

	enScreen.DrawImage(tilesImage.SubImage(image.Rect(sx, sy, sx+tileSpriteSize, sy+tileSpriteSize)).(*ebiten.Image), op)
}

func drawUnit(u *game.Unit, enScreen *ebiten.Image, cameraX, cameraY int) {
	screenPosition := u.Position.Mul(tileSize)
	x := screenPosition.X - float64(cameraX)
	y := screenPosition.Y - float64(cameraY)

	if u.Selected {
		col := color.RGBA{0, 255, 0, 255}
		ebitenutil.DrawRect(enScreen, x-selectedBorder, y-selectedBorder, float64(u.Size.X+(selectedBorder*2)), float64(u.Size.Y+(selectedBorder*2)), col)
	}

	ebitenutil.DrawRect(enScreen, x, y, float64(u.Size.X), float64(u.Size.Y), u.Color)
}

func getBackgroundColorImage(className string) *ebiten.Image {
	img, exists := backgroundImages[className]
	if exists {
		return img
	}
	switch className {
	case "grass":
		img = ebiten.NewImage(tileSize, tileSize)
		img.Fill(getColorFromHex("74CF45FF"))
	case "water":
		img = ebiten.NewImage(tileSize, tileSize)
		img.Fill(getColorFromHex("68A8C8FF"))
	case "sand":
		img = ebiten.NewImage(tileSize, tileSize)
		img.Fill(getColorFromHex("BA936BFF"))
	default:
		img = ebiten.NewImage(tileSize, tileSize)
		img.Fill(getColorFromHex("74CF45FF"))
	}
	backgroundImages[className] = img
	return img
}

func getColorFromHex(colorStr string) color.Color {
	b, err := hex.DecodeString(colorStr)
	if err != nil {
		log.Fatal(err)
	}

	return color.RGBA{b[0], b[1], b[2], b[3]}
}

func getTile(className string) int {
	switch className {
	case "plain1":
		return 0
	case "plain2":
		return 1
	case "plain3":
		return 2
	case "forest1":
		return 3
	case "forest2":
		return 4
	case "forest3":
		return 5
	case "sea1", "sea2", "sea3", "river1", "river2", "river3":
		return 99
	case "mountain1":
		return 10
	case "mountain2":
		return 11
	case "mountain3":
		return 12
	case "hill1", "hill2", "hill3":
		return 13
	case "lake1":
		return 18
	case "lake2":
		return 19
	case "lake3":
		return 20
	case "sand1", "sand2", "sand3":
		return 78
	default:
		return 0
	}
}
