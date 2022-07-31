package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"

	"github.com/tarm/serial"
)

// Config - data structure for the yaml config file information
type Config struct {
	WebPort int `yaml:"webport"` // which zone prefix
}

// type datasend struct {
// 	Data SpectroDataValues
// }

type SpectroDataValues struct {
	Serial int
	// Firmware     string
	// FirmwareDate string
	Red          int
	Green        int
	Blue         int
	XRespS       int
	YRespI       int
	IntRespM     int
	DeltaC       int
	ColorNo      int
	Group        int
	Trigger      int
	Temp         int
	RawRed       int
	RawGreen     int
	RawBlue      int
	MinRed       int
	MinGreen     int
	MinBlue      int
	MaxRed       int
	MaxGreen     int
	MaxBlue      int
	DoubParamSet int
}

func getInt(s []byte) int {
	var b [8]byte
	copy(b[8-len(s):], s)
	return int(binary.BigEndian.Uint64(b[:]))
}

func sendCmd(usedCmd []byte, sdv SpectroDataValues) SpectroDataValues {

	c := &serial.Config{
		Name:        "/dev/ttyUSB0",
		Baud:        115200,
		ReadTimeout: 100 * time.Millisecond,
		Size:        8,
		Parity:      'N',
		StopBits:    1}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	n, err := s.Write(usedCmd)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Sent %v bytes\n", n)

	var rHead []byte
	var rFrame []byte

	for n != 0 {
		buf := make([]byte, 1)
		n, err = s.Read(buf)
		if err == nil {
			rHead = append(rHead, buf...)
		}
	}
	// fmt.Printf("Received %v bytes\n", len(rHead))

	if len(rHead) > 8 {
		// fLength := getInt([]byte{rHead[5], rHead[4]})
		// fmt.Printf("Expected frame length %v bytes\n", fLength)
		rFrame = rHead[8:]
		rHead = rHead[:8]
	}

	// fmt.Printf("Header: %v\n", rHead)
	// fmt.Printf("Frame: %v\n", rFrame)

	if usedCmd[1] == 5 {
		sdv.Serial = getInt([]byte{rHead[3], rHead[2]})
	}

	// if usedCmd[1] == 7 {
	// 	sdv.Firmware = strings.TrimSpace(string((rFrame[:23])))
	// 	sdv.FirmwareDate = string(rFrame[23:])
	// }

	if usedCmd[1] == 8 {
		sdv.Red = getInt([]byte{rFrame[1], rFrame[0]})
		sdv.Green = getInt([]byte{rFrame[3], rFrame[2]})
		sdv.Blue = getInt([]byte{rFrame[5], rFrame[4]})
		sdv.XRespS = getInt([]byte{rFrame[7], rFrame[6]})
		sdv.YRespI = getInt([]byte{rFrame[9], rFrame[8]})
		sdv.IntRespM = getInt([]byte{rFrame[11], rFrame[10]})
		sdv.DeltaC = getInt([]byte{rFrame[13], rFrame[12]})
		sdv.ColorNo = getInt([]byte{rFrame[15], rFrame[14]})
		sdv.Group = getInt([]byte{rFrame[17], rFrame[16]})
		sdv.Trigger = getInt([]byte{rFrame[19], rFrame[18]})
		sdv.Temp = getInt([]byte{rFrame[21], rFrame[20]})
		sdv.RawRed = getInt([]byte{rFrame[23], rFrame[22]})
		sdv.RawGreen = getInt([]byte{rFrame[25], rFrame[24]})
		sdv.RawBlue = getInt([]byte{rFrame[27], rFrame[26]})
		sdv.MinRed = getInt([]byte{rFrame[29], rFrame[28]})
		sdv.MinGreen = getInt([]byte{rFrame[31], rFrame[30]})
		sdv.MinBlue = getInt([]byte{rFrame[33], rFrame[32]})
		sdv.MaxRed = getInt([]byte{rFrame[35], rFrame[34]})
		sdv.MaxGreen = getInt([]byte{rFrame[37], rFrame[36]})
		sdv.MaxBlue = getInt([]byte{rFrame[39], rFrame[38]})
		sdv.DoubParamSet = getInt([]byte{rFrame[41], rFrame[40]})
	}

	s.Close()

	return sdv
}

func main() {
	// create flag to get run type, if added run in test mode
	boolTest := flag.Bool("test", false, "a bool to show if code needs to run in test mode add '-test'")
	// CMD := flag.Int("cmd", 5, "an int representing the command to send to device")
	flag.Parse()

	// load config yaml file
	filename, _ := filepath.Abs("config.yml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// marshal config file into variable
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	// check to see if the '-test' flag has been added - if so print out results
	if !*boolTest {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = ioutil.Discard
	}

	r := gin.Default()

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.StaticFile("/favicon.ico", "./favicon.ico")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"data": "Colour Read",
		})
	})

	r.GET("/READ", func(c *gin.Context) {

		cmd5 := []byte{85, 5, 0, 0, 0, 0, 170, 60}
		cmd8 := []byte{85, 8, 0, 0, 0, 0, 170, 118}

		var SpDataVal = SpectroDataValues{}

		SpDataVal = sendCmd(cmd5, SpDataVal)
		SpDataVal = sendCmd(cmd8, SpDataVal)

		if *boolTest {
			fmt.Println(SpDataVal)
		}

		c.IndentedJSON(200, gin.H{
			"data": SpDataVal,
		})

	})

	if *boolTest {
		r.Run(fmt.Sprintf("127.0.0.1:%v", config.WebPort))
	} else {
		r.Run(fmt.Sprintf(":%v", config.WebPort))
	}
}
