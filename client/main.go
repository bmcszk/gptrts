package main

import (
	"crypto/md5"
	"errors"
	"image/color"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"

	"github.com/bmcszk/gptrts/pkg/game"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

var (
	grassImage, dirtImage *ebiten.Image
)

func init() {
	var err error
	grassImage, _, err = ebitenutil.NewImageFromFile("grass.png")
	if err != nil {
		log.Fatalf("Failed to load grass image: %v", err)
	}

	dirtImage, _, err = ebitenutil.NewImageFromFile("dirt.png")
	if err != nil {
		log.Fatalf("Failed to load dirt image: %v", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal(errors.New("agrument missing"))
	}
	name := os.Args[1]

	hash := md5.Sum([]byte(name))
	id, err := uuid.FromBytes(hash[:])
	if err != nil {
		log.Fatal(err)
	}
	playerId := game.PlayerIdType{UUID: id}
	u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	outgoingActions := make(chan game.Action, 10)
	dispatch := clientDispatch(outgoingActions)
	g := NewGame(dispatch)

	// Read messages from the server
	go func() {
		for {
			_, bytes, err := ws.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}
			action, err := game.UnmarshalAction(bytes)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("handle %s", action.GetType())
			if err := g.HandleAction(action); err != nil {
				log.Println(err)
			}
		}
	}()

	go func() {
		for action := range outgoingActions {
			log.Printf("dispatch %s", action.GetType())
			if err := ws.WriteJSON(action); err != nil {
				log.Println("write:", err)
			}
		}
	}()

	dispatch(game.PlayerInitAction{
		Type: game.PlayerInitActionType,
		Payload: game.Player{
			Id:    playerId,
			Name:  name,
			Color: nameToColor(name),
		},
	})

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("RTS Game")

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func clientDispatch(outgoing chan game.Action) game.DispatchFunc {
	return func(action game.Action) {
		outgoing <- action
	}
}

func nameToColor(name string) color.RGBA {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	switch name {
	case "red":
		return color.RGBA{255, 0, 0, 255}
	case "green":
		return color.RGBA{0, 255, 0, 255}
	case "blue":
		return color.RGBA{0, 0, 255, 255}
	case "yellow":
		return color.RGBA{255, 255, 0, 255}
	case "cyan":
		return color.RGBA{0, 255, 255, 255}
	case "purple":
		return color.RGBA{255, 0, 255, 255}
	default:
		return randomRGBA()
	}
}

// randomRGBA generates a random color in RGBA format
func randomRGBA() color.RGBA {
	return color.RGBA{
		R: uint8(rand.Intn(256)),
		G: uint8(rand.Intn(256)),
		B: uint8(rand.Intn(256)),
		A: 255, // Set alpha to 255 for opaque color
	}
}
