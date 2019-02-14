# Tandon Bridge HW Automator

This project seeks to automate the bulk compilation and running of CPP files submitted by Tandon Bridge Students for weekly HW assignments.

It is written in Golang and uses concurrency to run the code synchronously while building it (when two or more files are passed).

## Current Features

* Walk the "Download All" directory from NYU Clases to find the chosen student cpp hw
* Build the HW and save the binary in a subdir
* Copy cpp files into a subdir (will be removed if texteditor feature is implemented)
* Run the binary
  * kill unresponse running processes through stdin by using an escape character without killing the go process
* Start the build and run processes from a chosen student (ie. from somewhere other than the beginning of the list of students)
* Concurrent building and running

## TODO

* match both 'q' and 'Q' in the file name
* parse command line flags
* open textediter with student's code while running
* test whether the walking of dirs is in a deterministic order as to preserve consistency

### Maybe

* Record files with build errors in a list