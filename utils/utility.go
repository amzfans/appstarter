package utils

// if the channel is available, just send to the channel and return true.
// If not available, will ignore and return false.
func SendToNoBlockBoolChannel(dstChan chan bool, value bool) bool {
	select {
	case dstChan <- value:
		return true
	default:
		return false
	}
}
