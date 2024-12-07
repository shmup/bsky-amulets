package main

import "time"

type Stats struct {
	Posts         int
	Amulets       int
	TotalAmulets  int
	Rate          float64
	lastPostTimes []time.Time
}

func (s *Stats) updateRate(now, startTime time.Time) {
	s.lastPostTimes = append(s.lastPostTimes, now)

	timeSinceStart := now.Sub(startTime)
	if timeSinceStart < time.Minute {
		s.Rate = float64(len(s.lastPostTimes)) / timeSinceStart.Seconds()
	} else {
		cutoff := now.Add(-time.Minute)
		for i, t := range s.lastPostTimes {
			if t.After(cutoff) {
				s.lastPostTimes = s.lastPostTimes[i:]
				break
			}
		}
		s.Rate = float64(len(s.lastPostTimes)) / 60.0
	}
}
