package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
)

// updateState saves program's state and fraction term net weights.
func (p program) updateState() {
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModeDir)
	}
	p.saveState()
	p.flushTermData()
}

// loadState loads state from the last execution. If state load unsuccessful, the function returns non-nil error.
func (p *program) loadState() error {
	logger.Printf("INFO: Loading state from last execution")

	conffile, err := os.OpenFile("temp/config"+p.nstr+".gob", os.O_RDONLY, fs.ModePerm)
	if err != nil {
		logger.Printf("ERROR: Cannot load state of last execution: %v\n", err)
		return err
	}
	defer conffile.Close()

	newp := &program{}

	if err = gob.NewDecoder(conffile).Decode(newp); err != nil {
		logger.Printf("ERROR: Previous state decode unsuccessful: %v\n", err)
		return err
	}

	if newp.N != p.N && p.N != 0 {
		logger.Printf("ERROR: Config plane size and required plane size don't match\n")
		return errors.New("loadState: error: config plane size and required plane size don't match")
	} else {
		if p.WORKERS != 0 {
			newp.WORKERS = p.WORKERS
		}
		*p = *newp
		p.nstr = strconv.Itoa(int(newp.N))
	}

	logger.Printf("INFO: Loaded program state:\ncells: %d\nworkers: %d\nN: %d\nlastRect: %v\nlastDiag: %v\n", p.CELLS, p.WORKERS, p.N, p.LastRect, p.LastDiag)
	return nil
}

// saveState saves program's state.
func (p program) saveState() {
	logger.Println("INFO: Saving program's state")

	conffile, _ := os.OpenFile("temp/config"+p.nstr+".gob", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fs.ModePerm)

	logger.Printf("INFO: Saving program state:\ncells: %d\nworkers: %d\nN: %d\nlastRect: %v\nlastDiag: %v\n", p.CELLS, p.WORKERS, p.N, p.LastRect, p.LastDiag)
	if err := gob.NewEncoder(conffile).Encode(p); err != nil {
		logger.Printf("ERROR: Failed to save state: %v\n", err)
	}
}

// getTermData loads term net weights saved from last execution if it was stopped prematurely. If data load unsuccessful, the function returns non-nil error.
func (p *program) getTermData() error {
	logger.Println("INFO: Obtaining fraction term weights")

	resfile, err := os.OpenFile("temp/res"+p.nstr+".bin", os.O_RDONLY, fs.ModePerm)
	if err != nil {
		logger.Printf("ERROR: Couldn't open %s: %v\n", "res"+p.nstr+".bin", err)
		return err
	}
	defer resfile.Close()

	if err := gob.NewDecoder(resfile).Decode(&p.weights); err != nil {
		logger.Printf("ERROR: Couldn't read data from %s: %v\n", resfile.Name(), err)
		return err
	}
	return nil
}

// flushTermData saves continued fraction terms's net weights.
func (p program) flushTermData() {
	logger.Println("INFO: Saving obtained data")

	resfile, err := os.OpenFile("temp/res"+p.nstr+".bin", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fs.ModePerm)
	if err != nil {
		logger.Printf("ERROR: Couldn't open term data file for writing: %v\n", err)
		return
	}
	defer resfile.Close()

	if err := gob.NewEncoder(resfile).Encode(p.weights); err != nil {
		logger.Printf("ERROR: Failed to write data to the file: %v\n", err)
		return
	}
}

// saveFinalResults saves calculated term weights in human-readable form.
func (p program) saveFinalResults() {
	logger.Println("INFO: Saving final results")

	resfile, err := os.OpenFile("result_"+p.nstr+".dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.ModePerm)
	if err != nil {
		logger.Printf("ERROR: Couldn't open file for the result output: %v\n", err)
		return
	}
	defer resfile.Close()

	var buffer bytes.Buffer
	for key := 1; key <= int(p.N); key++ {
		buffer.WriteString(fmt.Sprintf("%d:%d\n", key, p.weights[key]))
	}

	writer := bufio.NewWriter(resfile)
	if _, err = writer.Write(buffer.Bytes()); err != nil {
		logger.Printf("ERROR: Couldn't write final results to file: %v\n", err)
		return
	}
	writer.Flush()
}

// clearFiles deletes config and term data files.
func (p program) clearFiles() {
	logger.Printf("INFO: Deleting redundant files: %s and %s\n", "res"+p.nstr+".bin", "config"+p.nstr+".gob")
	os.Remove("temp/res" + p.nstr + ".bin")
	os.Remove("temp/config" + p.nstr + ".gob")
}
