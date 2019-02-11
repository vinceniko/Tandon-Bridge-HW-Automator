package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// default vars
var (
	inDir    = "/Users/vincentscomputer/Library/Mobile Documents/com~apple~CloudDocs/Tandon/TA/c_compiler/test/hw4"
	compiler = "g++"
	bin      string
)

// StudentDir represents the hw directory
type StudentDir struct {
	Path string
	Dir  []os.FileInfo
	HWs  []*HW
	Next chan struct{}
}

// HW has fields for accessing HW for building and running
type HW struct {
	CppFile   string // path for the CppFile relativ to StudentDir.Dir
	BuildFile string // path for the BuildFile relative to StudentDir.Dir
	Build     *exec.Cmd
	Run       *exec.Cmd
}

// New creates inits a new StudentDir by assigning the hw dir path, reading the files of the dir, and creating an array of length n where n is the number of files in the dir
func New() *StudentDir {
	var err error

	sd := StudentDir{Path: inDir}

	sd.Dir, err = ioutil.ReadDir(inDir)
	if err != nil {
		log.Fatalln(err)
	}
	sd.HWs = make([]*HW, len(sd.Dir))
	sd.Next = make(chan struct{})

	return &sd
}

// BuildIt builds the cpp file located in the hw directory
func (hw *HW) BuildIt() {
	hw.BuildFile = path.Join(bin, strings.TrimSuffix(hw.CppFile, ".cpp"))
	hw.Build = exec.Command(compiler, hw.CppFile, "-o", hw.BuildFile)
	hw.Build.Dir = inDir
	err := hw.Build.Run()
	if err != nil {
		log.Fatalln(1, err)
	}
}

// RunIt runs the binary in bin that was created by build it
func (hw *HW) RunIt() {
	hw.Run = exec.Command(hw.BuildFile)
	hw.Run.Dir = inDir
	hw.Run.Stdout = os.Stdout
	hw.Run.Stderr = os.Stderr
	hw.Run.Stdin = os.Stdin
	err := hw.Run.Run()
	if err != nil {
		log.Fatalln(2, err)
	}
}

// Exec runs a goroutine which builds all the files in StudentDir independently from running the files
func (sd *StudentDir) Exec() {
	go func() {
		i := 0
		for _, file := range sd.Dir {
			if path.Ext(file.Name()) == ".cpp" {
				hw := &sd.HWs[i]
				if *hw == nil {
					*hw = &HW{CppFile: file.Name()}
				}

				(*hw).BuildIt()

				i++
				sd.Next <- struct{}{}
			}
		}
	}()
	i := 0
	for _, file := range sd.Dir {
		if path.Ext(file.Name()) == ".cpp" {
			select {
			case <-sd.Next:
				hw := &sd.HWs[i]
				if *hw == nil {
					*hw = &HW{CppFile: file.Name()}
				}
				fmt.Println("\n" + (*hw).CppFile)

				(*hw).RunIt()

				i++
			}
		}
	}
}

// CreateBinDir creates the binary directory in the folder where the hws are located if it isnt already there
func CreateBinDir() {
	var err error

	if info, _ := os.Stat(inDir); !info.IsDir() {
		err = os.Mkdir(inDir, 0666)
		if err != nil {
			log.Fatalln(err)
		}
	}

	bin = path.Join("bin")
	err = os.Mkdir(bin, 0666)
	if err != nil && err.(*os.PathError).Err.Error() != "file exists" {
		log.Fatalln(err)
	}
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

// init initializes the channel and creates the Bin directory
func init() {
	CreateBinDir()
}

func main() {
	sd := New()
	sd.Exec()

	GetNewLine()
}
