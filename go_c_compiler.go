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

// StudentDir represents the hw directory
type StudentDir struct {
	Path string
	Bin  string // Path is the main working directory
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
func New(inDir string, bin string) *StudentDir {
	var err error

	sd := StudentDir{Path: inDir, Bin: bin}

	sd.Dir, err = ioutil.ReadDir(inDir)
	if err != nil {
		log.Fatalln(err)
	}
	sd.HWs = make([]*HW, len(sd.Dir)-1) // sd.HWs will be less than or equal to the len of sd.Dir (whether there are other filetypes in the dir). -1 excludes the bin directory
	sd.Next = make(chan struct{})

	return &sd
}

// BuildIt builds the cpp file located in the hw directory
func (hw *HW) BuildIt(compiler, inDir, bin string) {
	hw.BuildFile = path.Join("bin", strings.TrimSuffix(hw.CppFile, ".cpp"))
	hw.Build = exec.Command(compiler, hw.CppFile, "-o", hw.BuildFile)
	hw.Build.Dir = inDir
	err := hw.Build.Run()
	if err != nil {
		log.Fatalln(1, err)
	}
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

// Exec runs a goroutine which builds all the files in StudentDir independently from running the files
func (sd *StudentDir) Exec(compiler string) {
	go func() { // this function builds the code and signals the channel so that the built files can be run
		i := 0
		for _, file := range sd.Dir {
			if path.Ext(file.Name()) == ".cpp" {
				hw := &sd.HWs[i]
				if *hw == nil {
					*hw = &HW{CppFile: file.Name()}
				}

				(*hw).BuildIt(compiler, sd.Path, sd.Bin)

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

				(*hw).RunIt(sd.Bin)

				i++
			}
		}
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
		inDir    = "/Users/vincentscomputer/Library/Mobile Documents/com~apple~CloudDocs/Tandon/TA/c_compiler/test/hw4"
		compiler = "g++"
	)

	bin := CreateBinDir(inDir)

	sd := New(inDir, bin)
	sd.Exec(compiler)

	//

	GetNewLine()
}
