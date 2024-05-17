package main

import (
	"fmt"
	"strconv"
)

// https://github.com/robsoncouto/arduino-songs

const NOTE_B0 = 31
const NOTE_C1 = 33
const NOTE_CS1 = 35
const NOTE_D1 = 37
const NOTE_DS1 = 39
const NOTE_E1 = 41
const NOTE_F1 = 44
const NOTE_FS1 = 46
const NOTE_G1 = 49
const NOTE_GS1 = 52
const NOTE_A1 = 55
const NOTE_AS1 = 58
const NOTE_B1 = 62
const NOTE_C2 = 65
const NOTE_CS2 = 69
const NOTE_D2 = 73
const NOTE_DS2 = 78
const NOTE_E2 = 82
const NOTE_F2 = 87
const NOTE_FS2 = 93
const NOTE_G2 = 98
const NOTE_GS2 = 104
const NOTE_A2 = 110
const NOTE_AS2 = 117
const NOTE_B2 = 123
const NOTE_C3 = 131
const NOTE_CS3 = 139
const NOTE_D3 = 147
const NOTE_DS3 = 156
const NOTE_E3 = 165
const NOTE_F3 = 175
const NOTE_FS3 = 185
const NOTE_G3 = 196
const NOTE_GS3 = 208
const NOTE_A3 = 220
const NOTE_AS3 = 233
const NOTE_B3 = 247
const NOTE_C4 = 262
const NOTE_CS4 = 277
const NOTE_D4 = 294
const NOTE_DS4 = 311
const NOTE_E4 = 330
const NOTE_F4 = 349
const NOTE_FS4 = 370
const NOTE_G4 = 392
const NOTE_GS4 = 415
const NOTE_A4 = 440
const NOTE_AS4 = 466
const NOTE_B4 = 494
const NOTE_C5 = 523
const NOTE_CS5 = 554
const NOTE_D5 = 587
const NOTE_DS5 = 622
const NOTE_E5 = 659
const NOTE_F5 = 698
const NOTE_FS5 = 740
const NOTE_G5 = 784
const NOTE_GS5 = 831
const NOTE_A5 = 880
const NOTE_AS5 = 932
const NOTE_B5 = 988
const NOTE_C6 = 1047
const NOTE_CS6 = 1109
const NOTE_D6 = 1175
const NOTE_DS6 = 1245
const NOTE_E6 = 1319
const NOTE_F6 = 1397
const NOTE_FS6 = 1480
const NOTE_G6 = 1568
const NOTE_GS6 = 1661
const NOTE_A6 = 1760
const NOTE_AS6 = 1865
const NOTE_B6 = 1976
const NOTE_C7 = 2093
const NOTE_CS7 = 2217
const NOTE_D7 = 2349
const NOTE_DS7 = 2489
const NOTE_E7 = 2637
const NOTE_F7 = 2794
const NOTE_FS7 = 2960
const NOTE_G7 = 3136
const NOTE_GS7 = 3322
const NOTE_A7 = 3520
const NOTE_AS7 = 3729
const NOTE_B7 = 3951
const NOTE_C8 = 4186
const NOTE_CS8 = 4435
const NOTE_D8 = 4699
const NOTE_DS8 = 4978
const REST = 0

// change this to make the song slower or faster
var tempo = 140

// notes of the moledy followed by the duration.
// a 4 means a quarter note, 8 an eighteenth , 16 sixteenth, so on
// !!negative numbers are used to represent dotted notes,
// so -4 means a dotted quarter note, that is, a quarter plus an eighteenth!!
var melodyCantinaBand = []int{
	// Cantina BAnd - Star wars
	// Score available at https://musescore.com/user/6795541/scores/1606876
	NOTE_B4, -4, NOTE_E5, -4, NOTE_B4, -4, NOTE_E5, -4,
	NOTE_B4, 8, NOTE_E5, -4, NOTE_B4, 8, REST, 8, NOTE_AS4, 8, NOTE_B4, 8,
	NOTE_B4, 8, NOTE_AS4, 8, NOTE_B4, 8, NOTE_A4, 8, REST, 8, NOTE_GS4, 8, NOTE_A4, 8, NOTE_G4, 8,
	NOTE_G4, 4, NOTE_E4, -2,
	NOTE_B4, -4, NOTE_E5, -4, NOTE_B4, -4, NOTE_E5, -4,
	NOTE_B4, 8, NOTE_E5, -4, NOTE_B4, 8, REST, 8, NOTE_AS4, 8, NOTE_B4, 8,

	NOTE_A4, -4, NOTE_A4, -4, NOTE_GS4, 8, NOTE_A4, -4,
	NOTE_D5, 8, NOTE_C5, -4, NOTE_B4, -4, NOTE_A4, -4,
	NOTE_B4, -4, NOTE_E5, -4, NOTE_B4, -4, NOTE_E5, -4,
	NOTE_B4, 8, NOTE_E5, -4, NOTE_B4, 8, REST, 8, NOTE_AS4, 8, NOTE_B4, 8,
	NOTE_D5, 4, NOTE_D5, -4, NOTE_B4, 8, NOTE_A4, -4,
	NOTE_G4, -4, NOTE_E4, -2,
	NOTE_E4, 2, NOTE_G4, 2,
	NOTE_B4, 2, NOTE_D5, 2,

	NOTE_F5, -4, NOTE_E5, -4, NOTE_AS4, 8, NOTE_AS4, 8, NOTE_B4, 4, NOTE_G4, 4,
}

var melodyImperialMarch = []int{

	// Dart Vader theme (Imperial March) - Star wars
	// Score available at https://musescore.com/user/202909/scores/1141521
	// The tenor saxophone part was used

	NOTE_A4, -4, NOTE_A4, -4, NOTE_A4, 16, NOTE_A4, 16, NOTE_A4, 16, NOTE_A4, 16, NOTE_F4, 8, REST, 8,
	NOTE_A4, -4, NOTE_A4, -4, NOTE_A4, 16, NOTE_A4, 16, NOTE_A4, 16, NOTE_A4, 16, NOTE_F4, 8, REST, 8,
	NOTE_A4, 4, NOTE_A4, 4, NOTE_A4, 4, NOTE_F4, -8, NOTE_C5, 16,

	NOTE_A4, 4, NOTE_F4, -8, NOTE_C5, 16, NOTE_A4, 2, //4
	NOTE_E5, 4, NOTE_E5, 4, NOTE_E5, 4, NOTE_F5, -8, NOTE_C5, 16,
	NOTE_A4, 4, NOTE_F4, -8, NOTE_C5, 16, NOTE_A4, 2,

	NOTE_A5, 4, NOTE_A4, -8, NOTE_A4, 16, NOTE_A5, 4, NOTE_GS5, -8, NOTE_G5, 16, //7
	NOTE_DS5, 16, NOTE_D5, 16, NOTE_DS5, 8, REST, 8, NOTE_A4, 8, NOTE_DS5, 4, NOTE_D5, -8, NOTE_CS5, 16,

	NOTE_C5, 16, NOTE_B4, 16, NOTE_C5, 16, REST, 8, NOTE_F4, 8, NOTE_GS4, 4, NOTE_F4, -8, NOTE_A4, -16, //9
	NOTE_C5, 4, NOTE_A4, -8, NOTE_C5, 16, NOTE_E5, 2,

	NOTE_A5, 4, NOTE_A4, -8, NOTE_A4, 16, NOTE_A5, 4, NOTE_GS5, -8, NOTE_G5, 16, //7
	NOTE_DS5, 16, NOTE_D5, 16, NOTE_DS5, 8, REST, 8, NOTE_A4, 8, NOTE_DS5, 4, NOTE_D5, -8, NOTE_CS5, 16,

	NOTE_C5, 16, NOTE_B4, 16, NOTE_C5, 16, REST, 8, NOTE_F4, 8, NOTE_GS4, 4, NOTE_F4, -8, NOTE_A4, -16, //9
	NOTE_A4, 4, NOTE_F4, -8, NOTE_C5, 16, NOTE_A4, 2,
}

var melodyTetris = []int{
	NOTE_E5, 4, NOTE_B4, 8, NOTE_C5, 8, NOTE_D5, 4, NOTE_C5, 8, NOTE_B4, 8,
	NOTE_A4, 4, NOTE_A4, 8, NOTE_C5, 8, NOTE_E5, 4, NOTE_D5, 8, NOTE_C5, 8,
	NOTE_B4, -4, NOTE_C5, 8, NOTE_D5, 4, NOTE_E5, 4,
	NOTE_C5, 4, NOTE_A4, 4, NOTE_A4, 8, NOTE_A4, 4, NOTE_B4, 8, NOTE_C5, 8,

	NOTE_D5, -4, NOTE_F5, 8, NOTE_A5, 4, NOTE_G5, 8, NOTE_F5, 8,
	NOTE_E5, -4, NOTE_C5, 8, NOTE_E5, 4, NOTE_D5, 8, NOTE_C5, 8,
	NOTE_B4, 4, NOTE_B4, 8, NOTE_C5, 8, NOTE_D5, 4, NOTE_E5, 4,
	NOTE_C5, 4, NOTE_A4, 4, NOTE_A4, 4, REST, 4,

	NOTE_E5, 4, NOTE_B4, 8, NOTE_C5, 8, NOTE_D5, 4, NOTE_C5, 8, NOTE_B4, 8,
	NOTE_A4, 4, NOTE_A4, 8, NOTE_C5, 8, NOTE_E5, 4, NOTE_D5, 8, NOTE_C5, 8,
	NOTE_B4, -4, NOTE_C5, 8, NOTE_D5, 4, NOTE_E5, 4,
	NOTE_C5, 4, NOTE_A4, 4, NOTE_A4, 8, NOTE_A4, 4, NOTE_B4, 8, NOTE_C5, 8,

	NOTE_D5, -4, NOTE_F5, 8, NOTE_A5, 4, NOTE_G5, 8, NOTE_F5, 8,
	NOTE_E5, -4, NOTE_C5, 8, NOTE_E5, 4, NOTE_D5, 8, NOTE_C5, 8,
	NOTE_B4, 4, NOTE_B4, 8, NOTE_C5, 8, NOTE_D5, 4, NOTE_E5, 4,
	NOTE_C5, 4, NOTE_A4, 4, NOTE_A4, 4, REST, 4,

	NOTE_E5, 2, NOTE_C5, 2,
	NOTE_D5, 2, NOTE_B4, 2,
	NOTE_C5, 2, NOTE_A4, 2,
	NOTE_GS4, 2, NOTE_B4, 4, REST, 8,
	NOTE_E5, 2, NOTE_C5, 2,
	NOTE_D5, 2, NOTE_B4, 2,
	NOTE_C5, 4, NOTE_E5, 4, NOTE_A5, 2,
	NOTE_GS5, 2,
}

func main() {

	melody := melodyTetris

	minHz := 9999
	maxHz := 0
	for i := 0; i < len(melody); i += 2 {
		minHz = min(minHz, melody[i])
		maxHz = max(maxHz, melody[i])
	}

	c := 0

	for i := 0; i < len(melody); i += 2 {

		note := melody[i]
		notelen := melody[i+1]

		notecount := 0

		switch notelen {
		case -2:
			notecount = 12
		case 2:
			notecount = 8
		case -4:
			notecount = 6
		case 4:
			notecount = 4
		case -8:
			notecount = 3
		case 8:
			notecount = 2
		case -16:
			notecount = 3
		case 16:
			notecount = 2
		default:
			panic("Invalid note length: " + strconv.Itoa(notelen))
		}

		for nci := 0; nci < notecount; nci++ {

			perc := float64(note-minHz) / float64(maxHz-minHz)

			if note == REST {
				fmt.Printf("%d, ", 0)
			} else {
				fmt.Printf("%d, ", int64(1000+perc*1500))
			}

			c++
			if c%24 == 0 {
				fmt.Printf("\n")
			}

		}

	}

}
