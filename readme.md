# Tandon Bridge HW Automator

This project seeks to automate the bulk compilation and running of CPP files submitted by Tandon Bridge Students for weekly HW assignments.

It is written in Golang and uses concurrency to run the code synchronously while building it (when two or more files are passed).

## Current Features

* CLI integration
* Walk the "Download All" directory from NYU Clases to find the chosen student cpp hw
* Build the HW and save the binary in a subdir
* Copy cpp files into a subdir (to make them easier to find; remove maybe)
* Run the binary
  * Kill an unresponsive running processes through stdin by using an escape character without killing the entire Go process
* Start the build and run processes from a chosen student (ie. from somewhere other than the beginning of the list of students)
* Concurrent building and running
* Choice of whether to build synchronously or only after finished running the previous program (allows the grader to edit the next program to be run before building it)

## Future Features

* Interactive commands
* Interactively step forwards and backwards through submissions
* Interactively restart execution

## Requirements

* tester.py
  * pexp
* Golang
* Submissions must be in their own subdirectory of the root project directory