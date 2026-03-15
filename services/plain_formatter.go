package services

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/muesli/termenv"
)

type PlainFormatter struct {
	w        io.Writer
	isTTY    bool
	maxWidth int
	profile  termenv.Profile
}

func NewPlainFormatter(w io.Writer, isTTY bool, maxWidth int) *PlainFormatter {
	profile := termenv.Ascii
	if isTTY {
		profile = termenv.ColorProfile()
	}
	return &PlainFormatter{w: w, isTTY: isTTY, maxWidth: maxWidth, profile: profile}
}

func (f *PlainFormatter) color(style termenv.ANSIColor) func(string) string {
	if !f.isTTY {
		return nil
	}
	return func(s string) string {
		return termenv.String(s).Foreground(f.profile.Color(strconv.Itoa(int(style)))).String()
	}
}

func (f *PlainFormatter) bold() func(string) string {
	if !f.isTTY {
		return nil
	}
	return func(s string) string {
		return termenv.String(s).Bold().String()
	}
}

func (f *PlainFormatter) dim() func(string) string {
	if !f.isTTY {
		return nil
	}
	return func(s string) string {
		return termenv.String(s).Faint().String()
	}
}

func (f *PlainFormatter) RenderStatus(entries []StatusEntry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(f.w, "No new activity since last check.")
		return err
	}

	if _, err := fmt.Fprintln(f.w, "\n📈 Status"); err != nil {
		return err
	}

	tp := tableprinter.New(f.w, f.isTTY, f.maxWidth)
	tp.AddHeader([]string{"Repo", "⭐", "🐛", "🔀", "🍴", "📦"})

	for _, e := range entries {
		tp.AddField(e.Repo, tableprinter.WithColor(f.bold()))

		if f.hasEvent(e.Events, "stars") && e.NewStars > 0 {
			tp.AddField(fmt.Sprintf("+%d", e.NewStars), tableprinter.WithColor(f.color(termenv.ANSIGreen)))
		} else {
			tp.AddField("", tableprinter.WithColor(f.dim()))
		}

		if f.hasEvent(e.Events, "issues") && e.NewIssues > 0 {
			tp.AddField(fmt.Sprintf("+%d", e.NewIssues), tableprinter.WithColor(f.color(termenv.ANSIGreen)))
		} else {
			tp.AddField("", tableprinter.WithColor(f.dim()))
		}

		if f.hasEvent(e.Events, "pull_requests") && e.NewPRs > 0 {
			tp.AddField(fmt.Sprintf("+%d", e.NewPRs), tableprinter.WithColor(f.color(termenv.ANSIGreen)))
		} else {
			tp.AddField("", tableprinter.WithColor(f.dim()))
		}

		if f.hasEvent(e.Events, "forks") && e.NewForks > 0 {
			tp.AddField(fmt.Sprintf("+%d", e.NewForks), tableprinter.WithColor(f.color(termenv.ANSIGreen)))
		} else {
			tp.AddField("", tableprinter.WithColor(f.dim()))
		}

		if f.hasEvent(e.Events, "releases") {
			if e.NewRelease {
				tp.AddField(e.ReleaseTag, tableprinter.WithColor(f.color(termenv.ANSIGreen)))
			} else if e.UnreleasedCount > 0 {
				tp.AddField(fmt.Sprintf("%d unreleased", e.UnreleasedCount), tableprinter.WithColor(f.color(termenv.ANSIYellow)))
			} else {
				tp.AddField("", tableprinter.WithColor(f.dim()))
			}
		} else {
			tp.AddField("", tableprinter.WithColor(f.dim()))
		}

		tp.EndRow()
	}

	return tp.Render()
}

func (f *PlainFormatter) RenderDashboard(result DashboardResult) error {
	if _, err := fmt.Fprintln(f.w, "\n📊 Dashboard"); err != nil {
		return err
	}

	tp := tableprinter.New(f.w, f.isTTY, f.maxWidth)
	tp.AddHeader([]string{"Repo", "⭐", "🐛", "🔀", "🍴", "📦", "Updated"})

	for _, e := range result.Repos {
		tp.AddField(e.Repo, tableprinter.WithColor(f.bold()))
		tp.AddField(fmt.Sprintf("%d", e.Stars))
		tp.AddField(fmt.Sprintf("%d", e.Issues))
		tp.AddField(fmt.Sprintf("%d", e.PullRequests))
		tp.AddField(fmt.Sprintf("%d", e.Forks))
		if e.LatestRelease != "" {
			release := e.LatestRelease
			if e.UnreleasedCount > 0 {
				release = fmt.Sprintf("%s +%d", e.LatestRelease, e.UnreleasedCount)
			}
			tp.AddField(release)
		} else {
			tp.AddField("—", tableprinter.WithColor(f.dim()))
		}
		tp.AddField(HumanizeAge(e.UpdatedAt))
		tp.EndRow()
	}

	if err := tp.Render(); err != nil {
		return err
	}

	totals := result.Totals
	if _, err := fmt.Fprintf(f.w, "\n📈 Totals: ⭐ %d  🐛 %d  🔀 %d  🍴 %d",
		totals.Stars, totals.Issues, totals.PRs, totals.Forks); err != nil {
		return err
	}
	if totals.NeedRelease > 0 {
		if _, err := fmt.Fprintf(f.w, "  📦 %d needs release", totals.NeedRelease); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(f.w)
	return err
}

func (f *PlainFormatter) RenderReleases(releases []ReleaseInfo) error {
	if _, err := fmt.Fprintln(f.w, "\n📦 Releases"); err != nil {
		return err
	}

	tp := tableprinter.New(f.w, f.isTTY, f.maxWidth)
	tp.AddHeader([]string{"Repo", "Release", "Age", "Unreleased", "Status"})

	needsRelease := 0
	upToDate := 0
	noReleases := 0

	for _, r := range releases {
		tp.AddField(r.Repo, tableprinter.WithColor(f.bold()))

		if r.LatestRelease == "" {
			tp.AddField("—", tableprinter.WithColor(f.dim()))
		} else {
			tp.AddField(r.LatestRelease)
		}

		tp.AddField(r.ReleaseAge)

		if r.UnreleasedCount > 0 {
			tp.AddField(fmt.Sprintf("%d commits", r.UnreleasedCount))
		} else if r.LatestRelease != "" {
			tp.AddField("0 commits")
		} else {
			tp.AddField("—", tableprinter.WithColor(f.dim()))
		}

		switch r.Status {
		case "needs release":
			tp.AddField(r.Status, tableprinter.WithColor(f.color(termenv.ANSIYellow)))
			needsRelease++
		case "up to date":
			tp.AddField(r.Status, tableprinter.WithColor(f.color(termenv.ANSIGreen)))
			upToDate++
		default:
			tp.AddField(r.Status, tableprinter.WithColor(f.dim()))
			noReleases++
		}

		tp.EndRow()
	}

	if err := tp.Render(); err != nil {
		return err
	}

	parts := []string{}
	if needsRelease > 0 {
		parts = append(parts, fmt.Sprintf("%d repos need a release", needsRelease))
	}
	if upToDate > 0 {
		parts = append(parts, fmt.Sprintf("%d up to date", upToDate))
	}
	if noReleases > 0 {
		parts = append(parts, fmt.Sprintf("%d no releases", noReleases))
	}
	if len(parts) > 0 {
		if _, err := fmt.Fprintf(f.w, "\n %s\n", strings.Join(parts, " · ")); err != nil {
			return err
		}
	}

	return nil
}

func (f *PlainFormatter) RenderFans(result FansResult) error {
	if len(result.Fans) == 0 {
		_, err := fmt.Fprintln(f.w, "No stargazers found.")
		return err
	}

	if _, err := fmt.Fprintln(f.w, "\n🌟 Fans"); err != nil {
		return err
	}

	tp := tableprinter.New(f.w, f.isTTY, f.maxWidth)
	tp.AddHeader([]string{"User", "⭐", "Repos"})

	for _, fan := range result.Fans {
		tp.AddField(fan.Login, tableprinter.WithColor(f.bold()))
		tp.AddField(fmt.Sprintf("%d", fan.Count))
		tp.AddField(strings.Join(fan.Repos, ", "), tableprinter.WithColor(f.dim()))
		tp.EndRow()
	}

	if err := tp.Render(); err != nil {
		return err
	}

	_, err := fmt.Fprintf(f.w, "\n %d fans · %d stars across all repos\n", result.TotalFans, result.TotalStars)
	return err
}

func (f *PlainFormatter) hasEvent(events []string, event string) bool {
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}
