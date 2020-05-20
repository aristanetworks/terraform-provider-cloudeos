// Copyright (c) 2019 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.
// Subject to Arista Networks, Inc.'s EULA.
// FOR INTERNAL USE ONLY. NOT FOR DISTRIBUTION.

package time

import (
	"testing"
	"time"
)

func TestSingleTimeCreation(t *testing.T) {
	start := time.Now()
	r, err := NewRequestTime(start)
	if err != nil {
		t.Fatal(err)
	}

	if r.IsRange() {
		t.Fatalf("should not have created a range")
	}

	s, err := r.GetStartTime()
	if err != nil {
		t.Fatalf("could not get start: %s", err)
	}

	if !s.Equal(start) {
		t.Errorf("did not get the time we inserted: %v but expected %v", s, start)
	}
}

func TestTimeRangeCreation(t *testing.T) {
	end := time.Now()
	start := end.AddDate(0, 0, -1) // this time yesterday
	r, err := NewRequestTimeRange(start, end)
	if err != nil {
		t.Fatal(err)
	}

	if !r.IsRange() {
		t.Fatalf("did not create range")
	}

	s, e, err := r.GetStartAndEnd()
	if err != nil {
		t.Fatalf("failed to get start/end: %s", err)
	}

	if !s.Equal(start) {
		t.Errorf("did not get start time we inserted: %v but expected %v", s, start)
	}

	if !e.Equal(end) {
		t.Errorf("did not get end time we inserted: %v but expected %v", e, end)
	}
}

func TestGetIndividualTimes(t *testing.T) {
	end := time.Now()
	start := end.AddDate(0, 0, -1) // this time yesterday

	//
	// single time
	//

	singleTime, err := NewRequestTime(start)
	if err != nil {
		t.Fatal(err)
	}

	s, err := singleTime.GetStartTime()
	if err != nil {
		t.Errorf("failed to get start: %s", err)
	}
	if !s.Equal(start) {
		t.Errorf("did not get start time we inserted: %v but expected %v", s, start)
	}

	e, err := singleTime.GetEndTime()
	if err == nil {
		t.Fatalf("should return error on end-time for single time")
	}

	//
	// range
	//

	rangeTime, err := NewRequestTimeRange(start, end)
	if err != nil {
		t.Fatal(err)
	}

	s, err = rangeTime.GetStartTime()
	if err != nil {
		t.Errorf("failed to get start: %s", err)
	}
	if !s.Equal(start) {
		t.Errorf("did not get start time we inserted: %v but expected %v", s, start)
	}

	e, err = rangeTime.GetEndTime()
	if err != nil {
		t.Errorf("failed to get end: %s", err)
	}
	if !e.Equal(end) {
		t.Errorf("did not get end time we inserted: %v but expected %v", e, end)
	}
}

func TestTimeRangeContains(outerT *testing.T) {
	end := time.Now()
	start := end.AddDate(0, 0, -1) // this time yesterday
	r, err := NewRequestTimeRange(start, end)
	if err != nil {
		outerT.Fatal(err)
	}

	if !r.IsRange() {
		outerT.Fatalf("did not create range")
	}

	tbl := []struct {
		Name      string
		Time      time.Time
		Inclusive bool
		IsWithin  bool
	}{
		{
			Name:      "Start_non_inclusive",
			Time:      start,
			Inclusive: false,
			IsWithin:  false,
		},
		{
			Name:      "Start_inclusive",
			Time:      start,
			Inclusive: true,
			IsWithin:  true,
		},
		{
			Name:      "End_non_inclusive",
			Time:      end,
			Inclusive: false,
			IsWithin:  false,
		},
		{
			Name:      "End_inclusive",
			Time:      end,
			Inclusive: true,
			IsWithin:  true,
		},
		{
			Name:      "Known_within_non_inclusive",
			Time:      start.Add(time.Minute * 10),
			Inclusive: false,
			IsWithin:  true,
		},
		{
			Name:      "Known_within_inclusive",
			Time:      start.Add(time.Minute * 10),
			Inclusive: true,
			IsWithin:  true,
		},
	}

	for _, test := range tbl {
		outerT.Run(test.Name, func(t *testing.T) {
			ok, err := r.Contains(test.Time, test.Inclusive)
			if err != nil {
				t.Fatalf("error while checking contains (inclusive=%v): %s", test.Inclusive, err)
			}

			if ok != test.IsWithin {
				t.Errorf("expected contained=%v for time %s in range (%s, %s)",
					test.IsWithin, test.Time, start, end)
			}
		})
	}
}
