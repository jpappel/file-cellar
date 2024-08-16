package storage

import "fmt"

type Stats struct {
	Uploaded   uint32
	Downloaded uint32
	Deleted    uint32
	Failed     uint32
}

type Statser interface {
	Stats() Stats
}

// Prints the usage statistics in the form "uploaded downloaded deleted Failed"
func (s Stats) String() string {
	return fmt.Sprint(s.Uploaded, s.Downloaded, s.Deleted, s.Failed)
}

func SumStats(s []Stats) Stats {
	total := Stats{}

	for _, v := range s {
		total.Uploaded += v.Uploaded
		total.Downloaded += v.Downloaded
		total.Deleted += v.Deleted
		total.Failed += v.Failed
	}

	return total
}
