package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	audioStreamURL := "https://ice-sov.musicradio.com/SmoothUKMP3"

	response, err := http.Get(audioStreamURL)

	if err != nil {
		fmt.Println("Error fetching audio stream: ", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: Received non-200 response:", response.StatusCode)
		return
	}

	streamer, format, err := mp3.Decode(response.Body)
	if err != nil {
		fmt.Println("Error formating audio stream: ", err)
	}

	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(streamer)

	select {}

}
