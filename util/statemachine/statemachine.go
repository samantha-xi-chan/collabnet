package statemachine

import (
	"fmt"
	"log"
)

type Event int
type State int

// Transition 表示状态转移规则
type Transition struct {
	CurrentState State
	Event        Event
	NextState    State
}

// StateMachine 表示状态机
type StateMachine struct {
	Transitions  []Transition
	CurrentState State
}

// NewStateMachine 创建一个新的状态机
func NewStateMachine(transitions []Transition, initialState State) *StateMachine {
	return &StateMachine{
		Transitions:  transitions,
		CurrentState: initialState,
	}
}

// HandleEvent 处理事件并进行状态转移
func (sm *StateMachine) HandleEvent(event Event) error {
	for _, transition := range sm.Transitions {
		if transition.CurrentState == sm.CurrentState && transition.Event == event {
			backup := transition.CurrentState
			sm.CurrentState = transition.NextState

			if backup != sm.CurrentState {
				log.Printf("[FSM] Change from %d to %d ( by evt %d ) \n", backup, sm.CurrentState, event)
			}

			return nil
		}
	}

	return fmt.Errorf("invalid transition for event %d in state %d", event, sm.CurrentState)
}
