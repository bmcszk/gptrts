package game

import (
	"encoding/json"
	"errors"

	"github.com/bmcszk/gptrts/pkg/world"
)

type ActionType string

const (
	PlayerInitActionType        ActionType = "PlayerInit"
	PlayerInitSuccessActionType ActionType = "PlayerInitSuccess"
	AddUnitActionType           ActionType = "AddUnit"
	MoveStartActionType         ActionType = "MoveStart"
	MoveStepActionType          ActionType = "MoveStep"
	MoveStopActionType          ActionType = "MoveStop"
	MapLoadActionType           ActionType = "MapLoad"
	MapLoadSuccessActionType    ActionType = "MapLoadSuccess"
)

type Action interface {
	GetType() ActionType
	GetPayload() any
}

type GenericAction[T any] struct {
	Type    ActionType
	Payload T
}

func (a GenericAction[T]) GetType() ActionType {
	return a.Type
}

func (a GenericAction[T]) GetPayload() any {
	return a.Payload
}

type PlayerInitAction = GenericAction[Player]

type PlayerInitSuccessAction = GenericAction[PlayerInitSuccessPayload]

type PlayerInitSuccessPayload struct {
	PlayerId PlayerIdType
	Map      Map
	Units    []Unit
	Players  []Player
}

type AddUnitAction = GenericAction[Unit]

type MoveStartAction = GenericAction[MoveStartPayload]

type MoveStartPayload struct {
	UnitId UnitIdType
	Point  PF
}

type MoveStepAction = GenericAction[MoveStepPayload]

type MoveStepPayload struct {
	UnitId   UnitIdType
	Position PF
	Path     []PF
	Step     int
}

type MoveStopAction = GenericAction[UnitIdType]

type MapLoadAction = GenericAction[MapLoadPayload]

type MapLoadPayload struct {
	world.WorldRequest
	PlayerId PlayerIdType
}

type MapLoadSuccessAction = GenericAction[MapLoadSuccessPayload]

type MapLoadSuccessPayload struct {
	world.WorldResponse
	PlayerId PlayerIdType
}

func UnmarshalAction(bytes []byte) (Action, error) {
	var msg GenericAction[any]
	if err := json.Unmarshal(bytes, &msg); err != nil {
		return nil, err
	}
	switch msg.Type {
	case PlayerInitActionType:
		var action PlayerInitAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case PlayerInitSuccessActionType:
		var action PlayerInitSuccessAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case AddUnitActionType:
		var action AddUnitAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case MoveStartActionType:
		var action MoveStartAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case MoveStepActionType:
		var action MoveStepAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case MoveStopActionType:
		var action MoveStopAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case MapLoadActionType:
		var action MapLoadAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	case MapLoadSuccessActionType:
		var action MapLoadSuccessAction
		if err := json.Unmarshal(bytes, &action); err != nil {
			return nil, err
		}
		return action, nil

	default:
		return nil, errors.New("action type unrecognized")
	}
}
