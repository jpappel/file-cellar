package storage

import "fmt"

// Track stats of file operations
type Stats struct {
	Uploaded   uint32 // number of uploaded files
	Downloaded uint32 // number of downloaded files
	Redirected uint32 // number of redirects, should be mutually exclusive from downloads
	Deleted    uint32 // number of deleted files
	Failed     uint32 // number of failures
}

// A summary of multiple Stats
type StatsSummary struct {
	Count   uint
	Minimum Stats
	Maximum Stats
	Average statsFloat
	Total   Stats
	StdDev  statsFloat
}

type statsFloat struct {
	Uploaded   float32
	Downloaded float32
	Redirected float32
	Deleted    float32
	Failed     float32
}

// Prints the usage statistics in the form "uploaded downloaded redirected deleted Failed"
func (s Stats) String() string {
	return fmt.Sprint(s.Uploaded, s.Downloaded, s.Redirected, s.Deleted, s.Failed)
}

func (s statsFloat) String() string {
	return fmt.Sprint(s.Uploaded, s.Downloaded, s.Redirected, s.Deleted, s.Failed)
}

func (s StatsSummary) String() string {
	return fmt.Sprintf("Count: %d\nMin%+v\nMax%+v\nAvg%+v\nTotal%+v\nStdDev%+v", s.Count, s.Minimum, s.Maximum, s.Average, s.Total, s.StdDev)
}

// Add one stats object to another
func (this Stats) Add(that Stats) {
	this.Uploaded += that.Uploaded
	this.Downloaded += that.Downloaded
	this.Redirected += that.Redirected
	this.Deleted += that.Deleted
	this.Failed += that.Failed
}

func (this Stats) Min(that Stats) {
	this.Uploaded = min(this.Uploaded, that.Uploaded)
	this.Downloaded = min(this.Downloaded, that.Downloaded)
	this.Redirected = min(this.Redirected, that.Redirected)
	this.Deleted = min(this.Deleted, that.Deleted)
	this.Failed = min(this.Failed, that.Failed)
}

func (this Stats) Max(that Stats) {
	this.Uploaded = max(this.Uploaded, that.Uploaded)
	this.Downloaded = max(this.Downloaded, that.Downloaded)
	this.Redirected = max(this.Redirected, that.Redirected)
	this.Deleted = max(this.Deleted, that.Deleted)
	this.Failed = max(this.Failed, that.Failed)
}

// Sum multiple Stats into a single Stats
func SumStats(s []Stats) Stats {
	total := Stats{}

	for _, v := range s {
		total.Add(v)
	}

	return total
}

// Compute a summary from multiple Stats
func Summary(stats []Stats) StatsSummary {
	s := StatsSummary{Count: uint(len(stats))}

	// JP: couldn't tell if there was a single pass way to compute everything so
	//     two passes are made.
	for _, v := range stats {
		s.Minimum.Min(v)
		s.Maximum.Max(v)
		s.Total.Add(v)
	}

	s.Average.Uploaded = float32(s.Total.Uploaded) / float32(s.Count)
	s.Average.Downloaded = float32(s.Total.Downloaded) / float32(s.Count)
	s.Average.Redirected = float32(s.Total.Redirected) / float32(s.Count)
	s.Average.Deleted = float32(s.Total.Deleted) / float32(s.Count)
	s.Average.Failed = float32(s.Total.Failed) / float32(s.Count)

	for _, v := range stats {
		s.StdDev.Uploaded += float32(v.Uploaded) - s.Average.Uploaded
		s.StdDev.Downloaded += float32(v.Downloaded) - s.Average.Downloaded
		s.StdDev.Redirected += float32(v.Redirected) - s.Average.Redirected
		s.StdDev.Deleted += float32(v.Deleted) - s.Average.Deleted
		s.StdDev.Failed += float32(v.Failed) - s.Average.Failed
	}

	s.StdDev.Uploaded /= float32(s.Count)
	s.StdDev.Downloaded /= float32(s.Count)
	s.StdDev.Redirected /= float32(s.Count)
	s.StdDev.Deleted /= float32(s.Count)
	s.StdDev.Failed /= float32(s.Count)

	return s
}
