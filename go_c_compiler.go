package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// StudentDir represents the hw directory
type StudentDir struct {
	Path  string
	Bin   string // Path is the main working directory
	HWs   []*HW
	Next  chan struct{}
	Close chan struct{}
}

// HW has fields for accessing HW for building and running
type HW struct {
	Name      string
	CppFile   string // path for the CppFile relativ to StudentDir.Dir
	BuildFile string // path for the BuildFile relative to StudentDir.Dir
	Build     *exec.Cmd
	Run       *exec.Cmd
}

// New creates inits a new StudentDir by assigning the hw dir path, reading the files of the dir, and creating an array of length n where n is the number of files in the dir
func New(compiler string, inDir string, bin string, q string) *StudentDir {
	var err error

	sd := StudentDir{Path: inDir, Bin: bin}
	sd.HWs = make([]*HW, 0)
	sd.Next = make(chan struct{})
	sd.Close = make(chan struct{})
	go func() {
		err = filepath.Walk(inDir, func(sPath string, info os.FileInfo, err error) error {
			if strings.Contains(info.Name(), q) && strings.Contains(info.Name(), "_") && path.Ext(info.Name()) == ".cpp" {
				hw := &HW{CppFile: sPath, Name: info.Name()}
				sd.HWs = append(sd.HWs, hw)

				(*hw).BuildIt(compiler, sd.Path, sd.Bin)

				sd.Next <- struct{}{}
			}

			return nil
		})
		if err != nil {
			log.Fatalln(err)
		}

		close(sd.Close)
	}()

	i := 0
	for true {
		select {
		case <-sd.Next:
			hw := &sd.HWs[i]
			fmt.Println("\n" + (*hw).Name)

			(*hw).RunIt(sd.Bin)

			i++
		case <-sd.Close:
			return nil
		}
	}

	return &sd
}

// BuildIt builds the cpp file located in the hw directory
func (hw *HW) BuildIt(compiler, inDir, bin string) {
	hw.BuildFile = path.Join(bin, strings.TrimSuffix(hw.Name, ".cpp"))
	hw.Build = exec.Command(compiler, hw.CppFile, "-o", hw.BuildFile)
	hw.Build.Dir = inDir
// ProcReadForwarder is used to forward read bytes to the intended reader unless it is a kill command such as 'q', which kills the underlying process
type ProcReadForwarder struct {
	KillChar byte
	Proc     **os.Process // a ptr to the field in exec.Command which points to a Process; since the field gets populated after creating PF
	In       io.Reader    // the reading source we forward to
}

/// Read allows ProcReadForwarder to implement the Reader interface
// it is used to scan for a escape character which kills the running process without killing the go process
// all other scanned bytes are forwarded to the internal reader as normal
func (pf ProcReadForwarder) Read(p []byte) (n int, err error) {
	if p[0] == pf.KillChar && p[1] == 10 { // if the bytes read == 'q'. 10 is the null character
		return 0, (*pf.Proc).Kill()
	}
	return pf.In.Read(p) // forward to the internal Reader
}

// RunIt runs the binary in bin that was created by build it
func (hw *HW) RunIt(bin string) {
	hw.Run = exec.Command(hw.BuildFile)
	hw.Run.Dir = hw.Build.Dir
	hw.Run.Stdout = os.Stdout
	hw.Run.Stderr = os.Stderr
	hw.Run.Stdin = os.Stdin
	err := hw.Run.Run()
	if err != nil {
		log.Fatalln(2, err)
	}
}

// CreateBinDir creates the binary directory in the folder where the hws are located if it isnt already there
func CreateBinDir(inDir string) string {
	var err error

	if info, _ := os.Stat(inDir); !info.IsDir() {
		err = os.Mkdir(inDir, 0666)
		if err != nil {
			log.Fatalln(err)
		}
	}

	bin := path.Join(inDir, "bin")
	err = os.Mkdir(bin, 0666)
	if err != nil && err.(*os.PathError).Err.Error() != "file exists" {
		log.Fatalln(err)
	}

	return bin
}

// GetNewLine checks for newline from StdIn
func GetNewLine() {
	bio := bufio.NewReader(os.Stdin)

	// in case you need a string which contains the newline
	_, err := bio.ReadString('\n')
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	// default vars
	var (
		inDir    = "/Users/vincentscomputer/Library/Mobile Documents/com~apple~CloudDocs/Tandon/TA/c_compiler/test/homework #4"
		compiler = "g++"
		q        = "q6"
	)

	bin := CreateBinDir(inDir)

	_ = New(compiler, inDir, bin, q)

	//

	GetNewLine()
}
