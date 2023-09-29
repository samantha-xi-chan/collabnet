package statemachine

import "fmt"

// Event 表示状态机的事件
type Event string

// State 表示状态机的状态
type State string

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
			sm.CurrentState = transition.NextState
			return nil
		}
	}
	return fmt.Errorf("Invalid transition for event %s in state %s", event, sm.CurrentState)
}
