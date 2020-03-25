package quickfix

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func sessionIDFilenamePrefix(s SessionID) string {
	sender := []string{s.SenderCompID}
	if s.SenderSubID != "" {
		sender = append(sender, s.SenderSubID)
	}
	if s.SenderLocationID != "" {
		sender = append(sender, s.SenderLocationID)
	}

	target := []string{s.TargetCompID}
	if s.TargetSubID != "" {
		target = append(target, s.TargetSubID)
	}
	if s.TargetLocationID != "" {
		target = append(target, s.TargetLocationID)
	}

	fname := []string{s.BeginString, strings.Join(sender, "_"), strings.Join(target, "_")}
	if s.Qualifier != "" {
		fname = append(fname, s.Qualifier)
	}
	return strings.Join(fname, "-")
}

// closeFile behaves like Close, except that no error is returned if the file does not exist
func closeFile(f *os.File) error {
	if f != nil {
		if err := f.Close(); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}

// removeFile behaves like os.Remove, except that no error is returned if the file does not exist
func removeFile(fname string) error {
	err := os.Remove(fname)
	if (err != nil) && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// openOrCreateFile opens a file for reading and writing, creating it if necessary
func openOrCreateFile(fname string, perm os.FileMode) (f *os.File, err error) {
	if f, err = os.OpenFile(fname, os.O_RDWR, perm); err != nil {
		if f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE, perm); err != nil {
			return nil, fmt.Errorf("error opening or creating file: %s: %s", fname, err.Error())
		}
	}
	return f, nil
}

// SeqnumFile represents a memory mapped file storing seqnums
type SeqnumFile struct {
	mmapfile *os.File
	length   int
	data     []byte //mmaped slice
	tmp      []byte //tmp buf to read/write to
}

// Read reads the seqnum
func (sqnf *SeqnumFile) Read() (int, error) {
	if sqnf.tmp == nil {
		return -1, fmt.Errorf("SeqnumFile is not init")
	}
	copy(sqnf.tmp[0:sqnf.length], sqnf.data[0:sqnf.length])
	seqnum, err := strconv.Atoi(string(sqnf.tmp))
	if err != nil {
		return -1, fmt.Errorf("Failed to read seqnum. ERR: %v", err)
	}
	return seqnum, nil
}

// Write writes the seqnum
func (sqnf *SeqnumFile) Write(seqnum int) error {
	seqnumstr := fmt.Sprintf("%019d", seqnum)
	copy(sqnf.data[0:], []byte(seqnumstr))
	return nil
}

// Init initializes the seqnum file to open or create the file at fname with length length
func (sqnf *SeqnumFile) Init(fname string, length int) error {
	var err error
	sqnf.mmapfile, err = openOrCreateFile(fname, 0660)
	if err != nil {
		return err
	}
	fi, err := sqnf.mmapfile.Stat()
	if err != nil {
		return err
	}

	// write byte array of length we want so the file is big enough to be written to
	if fi.Size() < int64(length) {
		sqnf.mmapfile.Write(make([]byte, length))
	}
	sqnf.data, err = syscall.Mmap(int(sqnf.mmapfile.Fd()), 0, syscall.Getpagesize(), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	sqnf.length = length
	sqnf.tmp = make([]byte, length)
	return nil
}

// Reset resets the seqnum file
func (sqnf *SeqnumFile) Reset() error {
	return sqnf.Write(1)
}

// Close closes the SeqnumFile
func (sqnf *SeqnumFile) Close() error {
	if sqnf.mmapfile == nil {
		return nil
	}
	if err := closeFile(sqnf.mmapfile); err != nil {
		return err
	}
	sqnf.mmapfile = nil
	return nil
}
