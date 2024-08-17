package storage

import "fmt"

type Stats struct {
	Uploaded   uint32
	Downloaded uint32
	Redirected uint32
	Deleted    uint32
	Failed     uint32
}

type Statser interface {
	Stats() Stats
}

// Prints the usage statistics in the form "uploaded downloaded redirected deleted Failed"
func (s Stats) String() string {
	return fmt.Sprint(s.Uploaded, s.Downloaded, s.Redirected, s.Deleted, s.Failed)
}

func (s Stats) Stats() Stats {
    return s
}

// Add one stats object to another
func (s1 Stats) Add(s2 Stats) {
	s1.Uploaded += s2.Uploaded
	s1.Downloaded += s2.Downloaded
	s1.Redirected += s2.Redirected
	s1.Deleted += s2.Deleted
	s1.Failed += s2.Failed
}

func SumStats(s []Stats) Stats {
	total := Stats{}

	for _, v := range s {
		total.Add(v)
	}

	return total
}
