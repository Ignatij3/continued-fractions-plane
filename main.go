package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"strconv"
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
func setupLogger(radius string) *os.File {
	_, err := os.Stat("logs")
	if os.IsNotExist(err) {
		os.Mkdir("logs", os.ModeDir)
	}

	logfile, _ := os.OpenFile("logs/"+radius+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, fs.ModePerm)
	logger = log.New(logfile, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
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
	NDEBUG = !NDEBUG
}

func main() {
	prg := &program{}
	parseArgs(prg)
	prg.nstr = strconv.Itoa(int(prg.N))
	defer setupLogger(prg.nstr).Close()

	logger.Printf("INFO: Parsed flags: n: %d; workers: %d; debug: %t\n", prg.N, prg.WORKERS, !NDEBUG)

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
	prg.clearFiles()
	if err := prg.saveFinalResults(); err != nil {
		log.Fatalf("FATAL: Couldn't write final results to file: %v\n Obtained data: %v\n", err, prg.weights)
	}
}
