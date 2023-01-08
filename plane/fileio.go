package plane

import (
	"encoding/gob"
	"errors"
	"io"
	"io/fs"
	"os"
	"strconv"
)

// TryToRestoreState tries to load configuration for plane of specific radius.
// If error is returned, it might be *PathError or custom error with no type.
func (p *Plane) TryToRestoreState() error {
	if err := p.loadState(); err != nil {
		return err
	}
	if err := p.getElementData(); err != nil {
		p.pcfg.LastLine = 0
		return err
	}

	return nil
}

// loadState loads state from the last execution, which includes number of workers and last processed line.
// If state load unsuccessful, the function returns non-nil error, which might be of type *PathError or decoding error.
// If plane has non-zero worker amount, it is not overwrited.
func (p *Plane) loadState() error {
	conffile, err := os.OpenFile("temp/config"+p.pcfg.rstr+".gob", os.O_RDONLY, fs.ModePerm)
	if err != nil {
		return err
	}
	defer conffile.Close()

	ncfg := planeConfig{}
	if err = gob.NewDecoder(conffile).Decode(&ncfg); err != nil {
		return err
	}

	if ncfg.Radius != p.pcfg.Radius && p.pcfg.Radius != 0 {
		return errors.New("loadState: error: config plane size and required plane size don't match")
	}

	if p.pcfg.Workers != 0 {
		ncfg.Workers = p.pcfg.Workers
	}
	p.pcfg = ncfg
	p.pcfg.rstr = strconv.Itoa(int(p.pcfg.Radius))

	return nil
}

// getElementData loads continued fraction element weights saved from last execution if it was stopped prematurely.
// If data load unsuccessful, the function returns non-nil error, which might be of type *PathError or decoding error.
func (p *Plane) getElementData() error {
	resfile, err := os.OpenFile("temp/res"+p.pcfg.rstr+".bin", os.O_RDONLY, fs.ModePerm)
	if err != nil {
		return err
	}
	defer resfile.Close()

	if err := gob.NewDecoder(resfile).Decode(&p.weights); err != nil {
		return err
	}
	return nil
}

// updateState saves program's state and continued fraction elements' weights.
// If error is returned, it might be of type *PathError, encoding error or custom error with no type.
func (p Plane) updateState() error {
	if p.running {
		return errors.New("plane: error: can't update state while calculations are running")
	}

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", os.ModeDir)
	}

	// save configuration
	conffile, _ := os.OpenFile("temp/config"+p.pcfg.rstr+".gob", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fs.ModePerm)
	if err := gob.NewEncoder(conffile).Encode(p.pcfg); err != nil {
		return err
	}

	// save data
	resfile, err := os.OpenFile("temp/res"+p.pcfg.rstr+".bin", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fs.ModePerm)
	if err != nil {
		return err
	}
	defer resfile.Close()

	if err := gob.NewEncoder(resfile).Encode(p.weights); err != nil {
		return err
	}

	return nil
}

// cleanup deletes all temporary files the module might have created.
func (p Plane) cleanup() {
	os.Remove("temp/res" + p.pcfg.rstr + ".bin")
	os.Remove("temp/config" + p.pcfg.rstr + ".gob")

	temp, _ := os.Open("temp")
	if _, err := temp.Readdirnames(1); err == io.EOF {
		os.Remove("temp")
	}
	temp.Close()
}
