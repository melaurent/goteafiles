This repo contains a Go implementation of TeaFiles

The implementation uses unsafe code for memory mapping of TeaFiles to memory.
The pointer returned by the MMapReader is unsafe and doesn't guarantee it 
is in the memory mapped region if the given item index is out of bound.


