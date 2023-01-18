package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"

	"./plane"
)

// initLogger binds logger to log file, returning the file to which the logger is writing.
func initLogger(radius int) *os.File {
	_, err := os.Stat("logs")
	if os.IsNotExist(err) {
		os.Mkdir("logs", os.ModeDir)
	}

	logfile, _ := os.OpenFile("logs/"+strconv.Itoa(radius)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, fs.ModePerm)
	log.SetOutput(logfile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	return logfile
}

// parseArgs parses and returns following command-line agruments:
// -r       - radius of the quarter
// -workers - amount of concurrent workers
func parseArgs() (uint, uint) {
	var r, workers uint

	flag.UintVar(&r, "r", 0, "radius of the quarter")
	flag.UintVar(&workers, "workers", 1, "amount of concurrent workers")

	flag.Parse()
	return r, workers
}

// saveResults writes results to a file, in human-readable form.
func saveResults(r int, weights []uint64) {
	resfile, err := os.OpenFile("result_"+strconv.Itoa(r)+".dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: Couldn't open file to save results to: %v\n", err)
	}
	defer resfile.Close()

	var buffer bytes.Buffer
	for key := 1; key <= r; key++ {
		buffer.WriteString(fmt.Sprintf("%d:%d\n", key, weights[key]))
	}

	writer := bufio.NewWriter(resfile)
	if _, err = writer.Write(buffer.Bytes()); err != nil {
		log.Fatalf("ERROR: Couldn't write data to save file: %v\n", err)
	}
	writer.Flush()
}

func main() {
	r, workers := parseArgs()

	initLogger(int(r))
	log.Printf("INFO: Parsed flags: r: %d; workers: %d;\n", r, workers)

	pln := plane.InitNewPlane(r, workers)

	if err := pln.TryToRestoreState(); err != nil {
		log.Printf("INFO: Couldn't restore state of last execution: %v\n", err)
	}

	pln.Start()
	<-pln.NotifyOnFinish()
	log.Printf("INFO: Calculations have stopped, finished state: %t\n", pln.IsFinished())
	if pln.IsFinished() {
		saveResults(int(r), pln.GetResults())
	}
}
