package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	NDEBUG bool = true
	logger *log.Logger
)

type program struct {
	WORKERS  uint
	N        uint
	nstr     string
	weights  []uint64
	LastLine uint
}

// setupLogger creates a new logger and binds it to file, returning the file to which the logger is writing.
func setupLogger() *os.File {
	_, err := os.Stat("logs")
	if os.IsNotExist(err) {
		os.Mkdir("logs", os.ModeDir)
	}

	logfile, _ := os.OpenFile("logs/"+time.Now().Format("02.01.2006_15.04.05.999999")+".log", os.O_CREATE, fs.ModePerm)
	logger = log.New(logfile, "", log.Ltime|log.Lmicroseconds|log.Lshortfile)
	return logfile
}

// parseArgs parses following command-line agruments:
// -restart - start new calculations
// -n       - size of the square plane
// -cells   - amount of the cells on one line, total amount of cells is approx. (cells^2)/2
// -workers - amount of the parallel workers
// -debug   - enables debug logs
func parseArgs(p *program) {
	flag.UintVar(&p.N, "n", 0, "size of the square plane")
	flag.UintVar(&p.WORKERS, "workers", 1, "amount of the parallel workers")
	flag.BoolVar(&NDEBUG, "debug", false, "enables debug logs")

	flag.Parse()

	logger.Printf("INFO: Parsed flags: n: %d; workers: %d; debug: %t\n", p.N, p.WORKERS, NDEBUG)
	NDEBUG = !NDEBUG
}

func main() {
	defer setupLogger().Close()
	prg := &program{}
	parseArgs(prg)
	prg.nstr = strconv.Itoa(int(prg.N))

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModeDir)
	}
	if err := prg.loadState(); err != nil {
		logger.Printf("INFO: Couldn't load state of last execution: %v\n", err)
	}
	if err := prg.getTermData(); err != nil {
		logger.Printf("INFO: Couldn't load fraction term data from last execution: %v\n", err)
	}

	prg.run()
	if err := prg.saveFinalResults(); err != nil {
		prg.clearFiles()
		log.Fatalf("FATAL: Couldn't write final results to file: %v\n Obtained data: %v\n", err, prg.weights)
	}
	prg.clearFiles()
}
