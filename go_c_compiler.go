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
			if startStudent != "" && foundDir == "" { // begin here
				dirs, err := ioutil.ReadDir(sd.Path)
				if err != nil {
					log.Fatalln(err)
				}
				for _, dir := range dirs {
					if strings.Contains(dir.Name(), startStudent) {
						foundDir = path.Join(sd.Path, dir.Name()) // find the dir name we need

						return nil
					}
				}

				log.Fatalln("NetID Not Found")
			} else if startStudent != "" && currPath != foundDir {
				if info, _ := os.Stat(currPath); currPath != sd.Path && info.IsDir() { // if currPath is not the root and is a dir
					return filepath.SkipDir
				}
				return nil // a non dir file, continue searching (if we have a file in the root dir, we don't want to skip)
			} else if startStudent != "" && currPath == foundDir { // located the student dir we need
				startStudent = ""
			} // executes the below once found

			if strings.Contains(info.Name(), q) && strings.Contains(info.Name(), "_") && path.Ext(info.Name()) == ".cpp" {
				hw := &HW{CppFile: currPath, Name: info.Name()}
				sd.HWs = append(sd.HWs, hw)

				hw.BuildIt(compiler, sd.Path, sd.BinDir, sd.CppDir)

				sd.Next <- hw // produce builds

				return filepath.SkipDir
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
		case hw := <-sd.Next: // consume builds
			fmt.Println("\n" + (*hw).Name)

			hw.RunIt(sd.BinDir)

			fmt.Println()
		case <-sd.Close: // nothing left to consume
			fmt.Println("\nDONE")
			return
		}
	}

// CopyCpp copies the cpp file to the cppdir
func CopyCpp(hw *HW, cppDir string) {
	src, err := os.Open(hw.CppFile)
	if err != nil {
		log.Fatalln(err)
		}
	defer src.Close()

	dst, err := os.Create(path.Join(cppDir, hw.Name))
	if err != nil {
		log.Fatalln(err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Fatalln(err)
	}
}

// BuildIt builds the cpp file located in the hw directory
func (hw *HW) BuildIt(compiler, inDir, bin string) {
	hw.BuildFile = path.Join(bin, strings.TrimSuffix(hw.Name, ".cpp"))
	hw.Build = exec.Command(compiler, hw.CppFile, "-o", hw.BuildFile)
	hw.Build.Dir = inDir
	go CopyCpp(hw, cppDir) // run in goroutine to prevent i/o delay
}
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

// CreateCppDir creates the binary directory in the folder where the hws are located if it isnt already there
func CreateCppDir(inDir string) string {
	var err error

	cppDir := path.Join(inDir, "cpps")
	if err = os.Mkdir(cppDir, 0666); err != nil && err.(*os.PathError).Err.Error() != "file exists" {
		log.Fatalln(err)
	}

	return cppDir
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
		student  = ""
	)

	bin := CreateBinDir(inDir)

	sd.Exec(compiler, student, q)

	//

	GetNewLine()
}
