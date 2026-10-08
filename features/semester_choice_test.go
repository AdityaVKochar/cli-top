package features

import (
	"cli-top/helpers"
	"cli-top/types"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSemesterCandidatesAutomaticNewestFirst(t *testing.T) {
	semesters := []types.Semester{
		{SemID: "oldest", SemName: "Oldest"},
		{SemID: "middle", SemName: "Middle"},
		{SemID: "newest", SemName: "Newest"},
	}

	got, err := semesterCandidates(semesters, 0)
	if err != nil {
		t.Fatalf("semesterCandidates returned an error: %v", err)
	}

	wantIDs := []string{"newest", "middle", "oldest"}
	if len(got) != len(wantIDs) {
		t.Fatalf("got %d candidates, want %d", len(got), len(wantIDs))
	}
	for i, wantID := range wantIDs {
		if got[i].SemID != wantID {
			t.Fatalf("candidate %d has ID %q, want %q", i, got[i].SemID, wantID)
		}
	}
}

func TestSemesterCandidatesPositiveChoiceSelectsOnlyDisplayedSemester(t *testing.T) {
	semesters := []types.Semester{
		{SemID: "first", SemName: "First displayed"},
		{SemID: "second", SemName: "Second displayed"},
		{SemID: "third", SemName: "Third displayed"},
	}

	got, err := semesterCandidates(semesters, 2)
	if err != nil {
		t.Fatalf("semesterCandidates returned an error: %v", err)
	}
	if len(got) != 1 || got[0].SemID != "second" {
		t.Fatalf("got %#v, want only the second displayed semester", got)
	}
}

func TestSemesterCandidatesRejectsInvalidChoice(t *testing.T) {
	semesters := []types.Semester{{SemID: "only", SemName: "Only"}}

	for _, choice := range []int{-1, 2} {
		t.Run(fmt.Sprintf("choice_%d", choice), func(t *testing.T) {
			if _, err := semesterCandidates(semesters, choice); err == nil {
				t.Fatalf("expected choice %d to be rejected", choice)
			}
		})
	}
}

var (
	courseSemesters = []string{"Fall Semester 2025-26", "Winter Semester 2025-26", "Fall Semester 2026-27"}
	octoberNow      = time.Date(2026, time.October, 8, 9, 0, 0, 0, time.UTC)
	februaryNow     = time.Date(2026, time.February, 8, 9, 0, 0, 0, time.UTC)
	decemberNow     = time.Date(2026, time.December, 8, 9, 0, 0, 0, time.UTC)
)

func TestChooseCourseSemesterPicksActiveSemester(t *testing.T) {
	t.Setenv("AUTO_SELECT_SEMESTER", "true")
	t.Setenv("CLI_TOP_PROXY_MODE", "")

	for _, tc := range []struct {
		now  time.Time
		want string
	}{
		{now: octoberNow, want: "Fall Semester 2026-27"},
		{now: februaryNow, want: "Winter Semester 2025-26"},
	} {
		got, err := chooseCourseSemester(courseSemesters, 0, tc.now)
		if err != nil {
			t.Fatalf("chooseCourseSemester(%s) returned an error: %v", tc.now.Month(), err)
		}
		if got != tc.want {
			t.Fatalf("chooseCourseSemester(%s) = %q, want %q", tc.now.Month(), got, tc.want)
		}
	}
}

func TestChooseCourseSemesterListsInTransitionMonths(t *testing.T) {
	t.Setenv("AUTO_SELECT_SEMESTER", "true")
	t.Setenv("CLI_TOP_PROXY_MODE", "")
	withStdin(t, "2\n")

	got, err := chooseCourseSemester(courseSemesters, 0, decemberNow)
	if err != nil {
		t.Fatalf("chooseCourseSemester returned an error: %v", err)
	}
	if got != "Winter Semester 2025-26" {
		t.Fatalf("got %q, want the semester picked from the list", got)
	}
}

func TestChooseCourseSemesterExplicitAndInvalidChoices(t *testing.T) {
	t.Setenv("AUTO_SELECT_SEMESTER", "true")
	got, err := chooseCourseSemester(courseSemesters, 1, octoberNow)
	if err != nil || got != "Fall Semester 2025-26" {
		t.Fatalf("chooseCourseSemester(1) = %q, %v; want first semester", got, err)
	}
	if _, err := chooseCourseSemester(courseSemesters, -1, octoberNow); err == nil {
		t.Fatal("expected a negative choice to be rejected")
	}
	if _, err := chooseCourseSemester(nil, 0, octoberNow); err == nil {
		t.Fatal("expected an empty semester list to be rejected")
	}
}

func TestChooseCourseSemesterListsWhenRequested(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  string
		list bool
	}{
		{name: "list flag", env: "true", list: true},
		{name: "auto-select off", env: "false", list: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AUTO_SELECT_SEMESTER", tc.env)
			t.Setenv("CLI_TOP_PROXY_MODE", "")
			helpers.ListSemesters = tc.list
			t.Cleanup(func() { helpers.ListSemesters = false })
			withStdin(t, "1\n")

			got, err := chooseCourseSemester(courseSemesters, 0, octoberNow)
			if err != nil {
				t.Fatalf("chooseCourseSemester returned an error: %v", err)
			}
			if got != "Fall Semester 2025-26" {
				t.Fatalf("got %q, want the semester picked from the list", got)
			}
		})
	}
}

func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	w.Close()
	original := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = original
		r.Close()
	})
}
