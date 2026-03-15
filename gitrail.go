package gitrail

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
)

const cmdName = "gitrail"

// Run the gitrail
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet(
		fmt.Sprintf("%s (v%s rev:%s)", cmdName, version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	ver := fs.Bool("version", false, "display version")
	since := fs.String("since", "", "start time (required)")
	until := fs.String("until", "", "end time (required)")
	dir := fs.String("C", "", "path to git repository (default: current directory)")
	branch := fs.String("branch", "", "target branch or revision (default: HEAD)")
	jsonOut := fs.Bool("json", false, "output as NDJSON")
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
		case Deleted:
			line = fmt.Sprintf("D\t%s", c.Path)
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

type jsonFileChange struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Status      string `json:"status"`
	Path        string `json:"path"`
	OldPath     string `json:"old_path,omitempty"`
}

func outputJSON(out io.Writer, result *Result) error {
	enc := json.NewEncoder(out)
	for _, c := range result.Changes {
		jc := jsonFileChange{
			Status: string(c.Status),
			Path:   c.Path,
		}
		switch c.Status {
		case Added:
			jc.To = result.To
		case Modified:
			jc.From = result.From
			jc.To = result.To
			jc.OldPath = c.OldPath
		case Deleted:
			jc.From = result.From
		}
		if err := enc.Encode(jc); err != nil {
			return err
		}
	}
	return nil
}
