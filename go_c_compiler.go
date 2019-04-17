package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// StudentDir represents the hw directory
type StudentDir struct {
	Path   string // master dir where hws are located
	CppDir string // where the student cpp files are stored
	BinDir string // where the binaries produced from the cpps are stored
	HWs    HWs
	Next   chan *HW // once built send over channel to "running" goroutine
	Seq    chan struct{}
	Close  chan struct{} // closed after no more hws are left to build and run
	Source string        // config var for changing behavior of dir walk
}

// HW has fields for accessing HW for building and running
type HW struct {
	Name      string // name of the binary
	CppFile   string // path for the CppFile relative to StudentDir.Path
	BuildFile string // path for the BuildFile relative to StudentDir.Path
	Build     *exec.Cmd
	Run       *exec.Cmd
}

// HWs is a slice of HW
type HWs []*HW

// New inits the StudentDir
func New(inDir, bin, cppDir, source string, seq bool) *StudentDir {
	sd := StudentDir{Path: inDir, BinDir: bin, CppDir: cppDir, Source: source}
	sd.HWs = make([]*HW, 0)
	sd.Next = make(chan *HW)
	if seq {
		sd.Seq = make(chan struct{}, 1)
		sd.Seq <- struct{}{}
	}
	sd.Close = make(chan struct{})

	return &sd
}

// WriteToHWChan writes HW chan
func (sd StudentDir) WriteToHWChan(hw *HW) {
	sd.Next <- hw
}

// Exec finds, runs, and builds
func (sd *StudentDir) Exec(compiler, startStudent, q string, times int, args ...string) {
	var err error

	go func() { // build producer
		err = filepath.Walk(sd.Path, func(currPath string, info os.FileInfo, err error) error {
			switch sd.Source {
			case "nyuclasses":
				if currPath == sd.BinDir || currPath == sd.CppDir {
					return filepath.SkipDir
				}

				foundDir := ""
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
			default:
				if currPath == sd.BinDir || currPath == sd.CppDir {
					return filepath.SkipDir
				}

				// general matcher
				if matchedFile := strings.Contains(info.Name(), startStudent); startStudent != "" && !matchedFile {
					return nil
				} else if matchedFile {
					startStudent = ""
				}
			}

			if q != "" { // if question specified
				if strings.Contains(info.Name(), q) && strings.Contains(info.Name(), "_") && path.Ext(info.Name()) == ".cpp" {
					hw := &HW{CppFile: currPath, Name: info.Name()}
					sd.HWs = append(sd.HWs, hw)

					hw.BuildIt(compiler, *sd)

					sd.WriteToHWChan(hw)
					return filepath.SkipDir
				}

				return nil // not the right file
			} else if path.Ext(info.Name()) == ".cpp" {
				hw := &HW{CppFile: currPath, Name: info.Name()}
				sd.HWs = append(sd.HWs, hw)

				hw.BuildIt(compiler, *sd)

				sd.WriteToHWChan(hw)
				return filepath.SkipDir
			}

			return nil
		})
		if err != nil {
			log.Fatalln(err)
		}

		close(sd.Close)
	}()

	for true { // build consumer
		select {
		case hw := <-sd.Next: // consume builds
			fmt.Println("\n" + (*hw).Name)

			for i := 0; i < times; i++ {
				hw.RunIt(*sd)
			}
			if sd.Seq != nil {
				sd.Seq <- struct{}{}
			}

			fmt.Println()
		case <-sd.Close: // nothing left to consume
			fmt.Println("\nDONE")
			return
		}
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
func (hw *HW) BuildIt(compiler string, sd StudentDir) {
	if sd.Seq != nil {
		select {
		case <-sd.Seq:
		}
	}
	hw.BuildFile = path.Join(sd.BinDir, strings.TrimSuffix(hw.Name, ".cpp"))
	hw.Build = exec.Command(compiler, hw.CppFile, "-o", hw.BuildFile)
	hw.Build.Dir = sd.Path
	if err := hw.Build.Run(); err != nil {
		fmt.Printf("Error Building %s; Err:%s\n", hw.Name, err) // TODO: save failed to compile files in a list
	}

	go CopyCpp(hw, sd.CppDir) // run in goroutine to prevent i/o delay
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
func (hw *HW) RunIt(sd StudentDir, args ...string) {
	hw.Run = exec.Command(hw.BuildFile, args...)
	hw.Run.Dir = hw.Build.Dir
	pf := ProcReadForwarder{'q', &hw.Run.Process, os.Stdin}
	hw.Run.Stdout = os.Stdout
	hw.Run.Stderr = os.Stderr
	hw.Run.Stdin = pf
	if err := hw.Run.Run(); err != nil {
		fmt.Printf("\nError Running %s; Err:%s\n", hw.Name, err)
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
	if err = os.Mkdir(bin, 0666); err != nil && err.(*os.PathError).Err.Error() != "file exists" {
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

	if _, err := bio.ReadString('\n'); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	var (
		inDir    = flag.String("inDir", "", "Dir with the student files")
		compiler = flag.String("compiler", "g++", "Compiler Command")
		q        = flag.String("q", "", "Question number in the form of \"q<n>\"")
		student  = flag.String("student", "", "Student to begin iteration with")
		source   = flag.String("source", "gradescope", "Student files download source")
		times    = flag.Int("times", 1, "Times to execute each program")
		seq      = flag.Bool("seq", false, "Whether to build sequentially")
		// args = flag.String("")
	)
	flag.Parse()
	for _, required := range []*string{inDir} {
		if *required == "" {
			fmt.Println("Error: A Required Var Must Be set")
		}
	}

	bin := CreateBinDir(*inDir)
	cppDir := CreateCppDir(*inDir)

	sd := New(*inDir, bin, cppDir, *source, *seq)
	sd.Exec(*compiler, *student, *q, *times, "--std=c++11")

	//

	GetNewLine() // press return once more to finish execution
}
