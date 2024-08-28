package predictor

import (
	"errors"
	"sync"
)

const maxPortNumber = 65535

type PortBitmap struct {
	bitmap []uint64
	mu     sync.Mutex
}

// NewPortBitmap initializes a PortBitmap with a size based on the maximum port number.
func NewPortBitmap() *PortBitmap {
	// Calculate the number of uint64 needed to cover the maximum port number
	numUint64 := (maxPortNumber + 1 + 63) / 64
	return &PortBitmap{
		bitmap: make([]uint64, numUint64),
	}
}

// SetPort sets the port as used in the bitmap.
func (pb *PortBitmap) SetPort(port int) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if port < 0 || port > maxPortNumber {
		return errors.New("port out of range")
	}
	pb.bitmap[port/64] |= 1 << (port % 64)
	return nil
}

// ClearPort clears the port as unused in the bitmap.
func (pb *PortBitmap) ClearPort(port int) error {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if port < 0 || port > maxPortNumber {
		return errors.New("port out of range")
	}
	pb.bitmap[port/64] &^= 1 << (port % 64)
	return nil
}

// IsPortSet checks if the port is used in the bitmap.
func (pb *PortBitmap) IsPortSet(port int) (bool, error) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if port < 0 || port > maxPortNumber {
		return false, errors.New("port out of range")
	}
	return pb.bitmap[port/64]&(1<<(port%64)) != 0, nil
}

// GetUsedPorts returns a list of all used ports.
func (pb *PortBitmap) GetUsedPorts() []int {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	var usedPorts []int
	for i, word := range pb.bitmap {
		for j := 0; j < 64; j++ {
			if word&(1<<j) != 0 {
				usedPorts = append(usedPorts, i*64+j)
			}
		}
	}
	return usedPorts
}
