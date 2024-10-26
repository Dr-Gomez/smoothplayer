package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Radio struct {
	Name     string
	DataType string
	Genre    string
	Url      string
}

type RadioCollection struct {
	Radios []Radio
}

func (rc *RadioCollection) AddRadio(name, dataType, genre, url string) {
	newRadio := Radio{
		Name:     name,
		DataType: dataType,
		Genre:    genre,
		Url:      url,
	}

	rc.Radios = append(rc.Radios, newRadio)
}

func (rc *RadioCollection) DisplayRadios() {

	separator := strings.Repeat("-", 150)

	fmt.Println()
	fmt.Println("The following radios are available at the moment:")

	for _, radio := range rc.Radios {
		fmt.Println(separator)
		fmt.Println()
		fmt.Printf("Radio Name: %s, Data Type: %s, Genre: %s, Url: %s\n", radio.Name, radio.DataType, radio.Genre, radio.Url)
		fmt.Println()
	}
	fmt.Print(separator)
	fmt.Println()

}

func clearTerminal() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func showHelp() {
	fmt.Print(`
Commands:
help:            Shows a list of all commands in the smoothplayer
clear:           Wipes out the terminal
stations:        Displays a list of all stations available
play (station):  Stream from the radio station selected
exit:            Quit the application 

`)
}

func omitFirstWord(input string) string {
	words := strings.Fields(input)
	if len(words) > 1 {
		return strings.ReplaceAll(strings.Join(words[1:], " "), " ", "")
	}

	return ""
}

// MP3 Stream

func streamMP3(audioStreamURL string, radioCollection RadioCollection) {
	speaker.Clear()

	for {
		response, err := http.Get(audioStreamURL)
		if err != nil {
			fmt.Println("Error fetching audio stream:", err)
			return
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			fmt.Println("Error: Received non-200 response:", response.StatusCode)
			return
		}

		streamer, format, err := mp3.Decode(response.Body)
		if err != nil {
			fmt.Println("Error decoding MP3 stream:", err)
			return
		}
		defer streamer.Close()

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		speaker.Play(streamer)

		scanner := bufio.NewScanner(os.Stdin)
		var musicURL string = ""

		for {
			fmt.Print("Enter command: ")
			if !scanner.Scan() {
				break
			}
			input := scanner.Text()

			if input == "exit" || input == "quit" || input == "q" {
				fmt.Println("Exiting the player...")
				return
			}

			if input == "clear" {
				clearTerminal()
			}

			if input == "help" {
				showHelp()
			}

			if input == "stations" {
				radioCollection.DisplayRadios()
			}

			if strings.HasPrefix(input, "play") {
				stationString := omitFirstWord(input)

				for _, station := range radioCollection.Radios {
					if strings.EqualFold(strings.ReplaceAll(station.Name, " ", ""), strings.ToLower(stationString)) {
						musicURL = station.Url + "MP3"
						break
					}
				}

				if musicURL != "" {
					fmt.Println("Now playing:", musicURL)
					break
				} else {
					fmt.Println("Station not found:", stationString)
				}
			}
		}

		if musicURL != "" {
			audioStreamURL = musicURL
		} else {
			fmt.Println("No music URL found. Replaying the original stream.")
		}
	}
}

func main() {
	baseSiteURL := "https://ice-sov.musicradio.com"
	audioStreamURL := "https://ice-sov.musicradio.com/ClassicFMMP3"

	// Base Website

	responseBase, errBase := http.Get(baseSiteURL)

	if errBase != nil {
		fmt.Println("Error fetching base site: ", errBase)
		return
	}

	defer responseBase.Body.Close()

	body, err := io.ReadAll(responseBase.Body)

	if err != nil {
		fmt.Println("Error reading response body: ", err)
		return
	}

	lines := strings.Split(string(body), "\n")

	radioCollection := RadioCollection{}

	var stream string = ""
	var genre string = ""
	var mediatype string = ""
	var url string = ""

	for _, line := range lines {
		trimLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimLine, `<td id="stream-title0">`) {
			stream = strings.Replace(strings.Replace(trimLine, `<td id="stream-title0">`, "", -1), `</td>`, "", -1)
		}

		if strings.HasPrefix(trimLine, `<td id="content-type0">`) {
			mediatype = strings.Replace(strings.Replace(trimLine, `<td id="content-type0">`, "", -1), `</td>`, "", -1)
		}

		if strings.HasPrefix(trimLine, `<td id="stream-genre0">`) {
			genre = strings.Replace(strings.Replace(trimLine, `<td id="stream-genre0">`, "", -1), `</td>`, "", -1)
		}

		if strings.HasPrefix(trimLine, `<td><a href="`) {
			trimLine = strings.Replace(trimLine, `<td><a href="`, "", -1)
			url = trimLine[:strings.Index(trimLine, `">`)]
		}

		if stream != "" && genre != "" && mediatype == "audio/mpeg" && url != "" {
			radioCollection.AddRadio(stream, mediatype, genre, url)
			stream = ""
			genre = ""
			mediatype = ""
			url = ""
		}

	}

	radioCollection.DisplayRadios()
	showHelp()

	streamMP3(audioStreamURL, radioCollection)

}
