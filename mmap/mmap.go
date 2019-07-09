//+build linux darwin

package mmap

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// MMapReader reads a memory-mapped tea file
type MMapReader struct {
	ItemCount int
	data      []byte
	ptr       uintptr
	size      int64
	itemSize  int64
}

// Close closes the reader.
func (r *MMapReader) Close() error {
	if r.data == nil {
		return nil
	}
	data := r.data
	r.data = nil
	runtime.SetFinalizer(r, nil)
	return syscall.Munmap(data)
}

// Len returns the number of items in the memory mapped region
func (r *MMapReader) Len() int {
	return int(r.size / r.itemSize)
}

// GetItem returns a point to the item at index idx
func (r *MMapReader) GetItem(idx int) unsafe.Pointer {
	return unsafe.Pointer(r.ptr + uintptr(idx * int(r.itemSize)))
}

// Open memory-maps the file for reading.
func Open(f *os.File, offset int64, size int64, itemSize int64) (*MMapReader, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	fSize := fi.Size()
	if fSize == 0 {
		return &MMapReader{}, nil
	}
	if fSize < 0 {
		return nil, fmt.Errorf("mmap: file has negative size")
	}
	if fSize != int64(int(fSize)) {
		return nil, fmt.Errorf("mmap: file is too large")
	}
	if size + offset > fSize {
		return nil, fmt.Errorf("mmap: size is too large")
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	ptr := uintptr(unsafe.Pointer(&data[0])) + uintptr(offset)
	r := &MMapReader{
		ItemCount: int(size / itemSize),
		data: data,
		ptr: ptr,
		size: size,
		itemSize: itemSize}

	runtime.SetFinalizer(r, (*MMapReader).Close)
	return r, nil
}
