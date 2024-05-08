package main

import (
	"fmt"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"gogs.mikescher.com/BlackForestBytes/goext/timeext"
	"strings"
	"time"
)

type Action string //@enum:type

const (
	ActionReset          Action = "RESET"
	ActionLamp           Action = "LAMP"
	ActionBuzzer1        Action = "BUZZER_1"
	ActionBuzzer2        Action = "BUZZER_2"
	ActionBuzzer3        Action = "BUZZER_3"
	ActionPWMBuzzer      Action = "BUZZER_PWM"
	ActionPWMBuzzerFunc  Action = "BUZZER_PWM_FUNC"
	ActionPWMBuzzerNotes Action = "BUZZER_PWM_NOTES"
)

type PWMFunction string //@enum:type

const (
	PWMFunctionSinus    PWMFunction = "SINUS"
	PWMFunctionTriangle PWMFunction = "TRIANGLE"
	PWMFunctionSawtooth PWMFunction = "SAWTOOTH"
	PWMFunctionSquare   PWMFunction = "SQUARE"
)

type Command struct {
	ID string `json:"-"`

	Date     time.Time  `json:"-"`
	Status   string     `json:"-"`
	Executed *time.Time `json:"-"`

	Action Action `json:"action"` // [ * ]
	Delay  int    `json:"delay"`  // [ * ]

	Duration int `json:"duration"` // [ * / BUZZER_PWM_NOTES ]

	// ### BUZZER_PWM

	Frequency int `json:"frequency"`

	// ### BUZZER_PWM_FUNC

	FrequencyMin int         `json:"frequencyMin"`
	FrequencyMax int         `json:"frequencyMax"`
	Func         PWMFunction `json:"func"`
	Period       int         `json:"period"`

	// ### BUZZER_PWM_NOTES

	NoteLength int   `json:"noteLength"`
	Notes      []int `json:"notes"`
}

func (c Command) Valid() (string, bool) {
	switch c.Action {
	case ActionReset:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		return "", true
	case ActionLamp:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if c.Duration <= 0 {
			return "Duration <= 0", false
		}
		return "", true
	case ActionBuzzer1:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if c.Duration <= 0 {
			return "Duration <= 0", false
		}
		return "", true

	case ActionBuzzer2:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if c.Duration <= 0 {
			return "Duration <= 0", false
		}
		return "", true

	case ActionBuzzer3:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if c.Duration <= 0 {
			return "Duration <= 0", false
		}
		return "", true

	case ActionPWMBuzzer:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if c.Duration <= 0 {
			return "Duration <= 0", false
		}
		if c.Frequency < 1000 {
			return "Frequency < 1000", false
		}
		if c.Frequency > 3000 {
			return "Frequency > 3000", false
		}
		return "", true

	case ActionPWMBuzzerFunc:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if c.Duration <= 0 {
			return "Duration < 0", false
		}
		if !c.Func.Valid() {
			return fmt.Sprintf("Invalid func: %s", c.Func), false
		}
		if c.FrequencyMin < 1000 {
			return "FrequencyMin < 1000", false
		}
		if c.FrequencyMax > 3000 {
			return "FrequencyMax > 3000", false
		}
		return "", true

	case ActionPWMBuzzerNotes:
		if c.Delay < 0 {
			return "Delay < 0", false
		}
		if len(c.Notes) == 0 {
			return "len(c.Notes) == 0", false
		}
		if c.NoteLength <= 0 {
			return "Period <= 0", false
		}
		for i, note := range c.Notes {
			if note < 1000 && note != 0 {
				return fmt.Sprintf("Notes[%d] < 1000 && Notes[%d] != 0", i, i), false
			}
			if note > 3000 {
				return fmt.Sprintf("Notes[%d] > 3000 && Notes[%d] != 0", i, i), false
			}
		}
		return "", true

	default:
		return fmt.Sprintf("Invalid action: %s", c.Action), false
	}
}

func (c Command) String() string {
	switch c.Action {
	case ActionReset:
		return fmt.Sprintf("%s;%s;delay=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay)
	case ActionLamp:
		return fmt.Sprintf("%s;%s;delay=%d;duration=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.Duration)
	case ActionBuzzer1:
		return fmt.Sprintf("%s;%s;delay=%d;duration=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.Duration)
	case ActionBuzzer2:
		return fmt.Sprintf("%s;%s;delay=%d;duration=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.Duration)
	case ActionBuzzer3:
		return fmt.Sprintf("%s;%s;delay=%d;duration=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.Duration)
	case ActionPWMBuzzer:
		return fmt.Sprintf("%s;%s;delay=%d;duration=%d;frequency=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.Duration, c.Frequency)
	case ActionPWMBuzzerFunc:
		return fmt.Sprintf("%s;%s;delay=%d;duration=%d;frequency_min=%d;frequency_max=%d;func=%s;period=%d", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.Duration, c.FrequencyMin, c.FrequencyMax, c.Func, c.Period)
	case ActionPWMBuzzerNotes:
		return fmt.Sprintf("%s;%s;delay=%d;note_length=%d;notes=%s", c.Date.In(timeext.TimezoneBerlin).Format(time.RFC3339), c.Action, c.Delay, c.NoteLength, strings.Join(langext.ArrMap(c.Notes, func(v int) string { return fmt.Sprintf("%d", v) }), ":"))
	default:
		return "ERROR"
	}
}
