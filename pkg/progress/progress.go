package progress

import "github.com/cheggaaa/pb/v3"

func StartProgress(total int) *pb.ProgressBar {
	return pb.StartNew(total)
}
