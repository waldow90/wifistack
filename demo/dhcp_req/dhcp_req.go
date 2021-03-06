package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/unixpickle/gofi"
	"github.com/unixpickle/wifistack"
	"github.com/unixpickle/wifistack/frames"
)

const Timeout = time.Second * 5

func main() {
	interfaceName, err := gofi.DefaultInterfaceName()
	if err != nil {
		log.Fatalln("no default interface:", err)
	}
	handle, err := gofi.NewHandle(interfaceName)
	if err != nil {
		log.Fatalln("could not open handle to "+interfaceName+":", err)
	}
	defer handle.Close()

	fmt.Println("BSS Descriptions:")

	stream := wifistack.NewRawStream(handle)
	scanRes, _ := wifistack.ScanNetworks(stream)
	descriptions := []frames.BSSDescription{}
	for desc := range scanRes {
		fmt.Println(len(descriptions), "-", desc.BSSID, desc.SSID)
		descriptions = append(descriptions, desc)
	}

	fmt.Print("Pick a number from the list: ")
	choice := readChoice()
	if choice < 0 || choice >= len(descriptions) {
		log.Fatalln("choice out of bounds.")
	}

	handshaker := wifistack.Handshaker{
		Stream: stream,
		Client: frames.MAC{0, 1, 2, 3, 4, 5},
		BSS:    descriptions[choice],
	}
	if err := handshaker.HandshakeOpen(time.Second * 5); err != nil {
		log.Fatalln("handshake failed:", err)
	} else {
		log.Println("handshake successful!")
	}

	msduConfig := wifistack.OpenMSDUStreamConfig{
		FragmentThreshold: 1000,
		DataRate:          2,
		BSSID:             handshaker.BSS.BSSID,
		Client:            handshaker.Client,
		Stream:            stream,
	}
	msduStream := wifistack.NewOpenMSDUStream(msduConfig)
	msduStream.Outgoing() <- wifistack.MSDU{
		Remote:  frames.MAC{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		Payload: []byte("\xAA\xAA\x03\x00\x00\x00\x08\x00\x45\x00\x01\x67\x85\xF7\x00\x00\x40\x11\xF3\x8F\x00\x00\x00\x00\xFF\xFF\xFF\xFF\x00\x44\x00\x43\x01\x53\xA9\xC5\x01\x01\x06\x00\x5F\xB0\xC2\x5F\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x24\xF5\xAA\x28\x2E\xE4\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x63\x82\x53\x63\x35\x01\x03\x3D\x07\x01\x24\xF5\xAA\x28\x2E\xE4\x32\x04\xAC\x14\x14\x14\x36\x04\xAC\x14\x14\x01\x39\x02\x05\xDC\x3C\x28\x64\x68\x63\x70\x63\x64\x2D\x36\x2E\x38\x2E\x32\x3A\x4C\x69\x6E\x75\x78\x2D\x33\x2E\x38\x2E\x31\x31\x3A\x61\x72\x6D\x76\x37\x6C\x3A\x53\x41\x4D\x53\x55\x4E\x47\x91\x01\x01\x37\x0F\x01\x79\x21\x03\x06\x0C\x0F\x1A\x1C\x33\x36\x3A\x3B\x77\xFC\xFF"),
	}
	for msdu := range msduStream.Incoming() {
		fmt.Println("got MSDU:", msdu)
	}
}

func readChoice() int {
	s := ""
	for {
		b := make([]byte, 1)
		if _, err := os.Stdin.Read(b); err != nil {
			log.Fatalln(err)
		}
		if b[0] == '\n' {
			break
		} else if b[0] == '\r' {
			continue
		} else {
			s += string(b)
		}
	}

	num, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalln("invalid number:", s)
	}
	return num
}
