package gitrail

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/Songmu/skillsmith"
	"github.com/itchyny/gojq"
)

//go:embed skills
var skillsFS embed.FS

const cmdName = "gitrail"

// Run the gitrail
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	if len(argv) > 0 && argv[0] == "skills" {
		s, err := skillsmith.New(cmdName, version, skillsFS)
		if err != nil {
			return err
		}
		s.OutWriter = outStream
		s.ErrWriter = errStream
		return s.Run(ctx, argv[1:])
	}
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [options] [-- pathspec...]\n\n", cmdName)
		fmt.Fprintf(fs.Output(), "Subcommands:\n")
		fmt.Fprintf(fs.Output(), "  skills    manage and distribute agent skills\n\n")
		fmt.Fprintf(fs.Output(), "Options:\n")
		fs.PrintDefaults()
	}
	ver := fs.Bool("version", false, "display version")
	since := fs.String("since", "", "start time (required)")
	until := fs.String("until", "", "end time (required)")
	dir := fs.String("C", "", "path to git repository (default: current directory)")
	branch := fs.String("branch", "", "target branch or revision (default: HEAD)")
	jsonOut := fs.Bool("json", false, "output as NDJSON")
	jqFilter := fs.String("jq", "", "filter JSON output using a jq expression")
	rawOutput := fs.Bool("r", false, "output raw strings with --jq")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}
	if *since == "" {
		fs.Usage()
		return fmt.Errorf("--since is required")
	}
	if *until == "" {
		fs.Usage()
		return fmt.Errorf("--until is required")
	}
	if *rawOutput && *jqFilter == "" {
		fs.Usage()
		return fmt.Errorf("-r requires --jq")
	}

	var jqCode *gojq.Code
	if *jqFilter != "" {
		query, err := gojq.Parse(*jqFilter)
		if err != nil {
			return fmt.Errorf("parse --jq expression: %w", err)
		}
		jqCode, err = gojq.Compile(query)
		if err != nil {
			return fmt.Errorf("compile --jq expression: %w", err)
		}
	}

	result, err := trail(ctx, &trailOpts{
		Dir:       *dir,
		Since:     *since,
		Until:     *until,
		Branch:    *branch,
		Pathspecs: fs.Args(),
	}, errStream)
	if err != nil {
		return err
	}

	if jqCode != nil {
		return outputJQ(outStream, result, jqCode, *rawOutput)
	}
	if *jsonOut {
		return outputJSON(outStream, result)
	}
	return outputText(outStream, result)
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "%s v%s (rev:%s)\n", cmdName, version, revision)
	return err
}

func outputText(out io.Writer, result *Result) error {
	_, err := fmt.Fprintf(out, "%s..%s\n", result.From, result.To)
	if err != nil {
		return err
	}
	if len(result.Changes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	for _, c := range result.Changes {
		var line string
		switch c.Status {
		case Added:
			line = fmt.Sprintf("A\t%s", c.Path)
		case Modified:
			if c.OldPath != "" {
				line = fmt.Sprintf("M\t%s\t%s", c.Path, c.OldPath)
			} else {
				line = fmt.Sprintf("M\t%s", c.Path)
			}
		case Renamed:
			line = fmt.Sprintf("R\t%s\t%s", c.Path, c.OldPath)
		case Deleted:
			line = fmt.Sprintf("D\t%s", c.Path)
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func outputJSON(out io.Writer, result *Result) error {
	enc := json.NewEncoder(out)
	for _, c := range result.Changes {
		jc := newJSONFileChange(result, c)
		if err := enc.Encode(jc); err != nil {
			return err
		}
	}
	return nil
}

func outputJQ(out io.Writer, result *Result, code *gojq.Code, raw bool) error {
	for _, c := range result.Changes {
		iter := code.Run(newJSONFileChange(result, c))
		for {
			value, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := value.(error); ok {
				return fmt.Errorf("run --jq expression: %w", err)
			}
			if raw {
				if s, ok := value.(string); ok {
					if _, err := fmt.Fprintln(out, s); err != nil {
						return err
					}
					continue
				}
			}
			b, err := gojq.Marshal(value)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(out, string(b)); err != nil {
				return err
			}
		}
	}
	return nil
}

func newJSONFileChange(result *Result, c FileChange) map[string]any {
	value := map[string]any{
		"status": string(c.Status),
		"path":   c.Path,
	}
	switch c.Status {
	case Added:
		value["to"] = result.To
	case Modified:
		value["from"] = result.From
		value["to"] = result.To
		if c.OldPath != "" {
			value["old_path"] = c.OldPath
		}
	case Renamed:
		value["from"] = result.From
		value["to"] = result.To
		if c.OldPath != "" {
			value["old_path"] = c.OldPath
		}
	case Deleted:
		value["from"] = result.From
	}
	return value
}
