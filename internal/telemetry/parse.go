package telemetry

import (
    "fmt"
    "math"
    "strconv"
    "strings"
)

type Report struct {
    Version int
    ID      string
    Seq     uint64
    Conns   int
    RPS     float64
}

const maxLineLen = 512

func Parse(line string) (Report, error) {
    if len(line) > maxLineLen {
        return Report{}, fmt.Errorf("line too long: %d bytes", len(line))
    }

    fields := make(map[string]string)
    for _, tok := range strings.Fields(line) {
        key, val, ok := strings.Cut(tok, "=")
        if !ok {
            return Report{}, fmt.Errorf("token %q has no '='", tok)
        }
        if _, dup := fields[key]; dup {
            return Report{}, fmt.Errorf("duplicate key %q", key)
        }
        fields[key] = val
    }

    var r Report

    v, ok := fields["v"]
    if !ok {
        return Report{}, fmt.Errorf("missing v")
    }
    n, err := strconv.Atoi(v)
    if err != nil || n != 1 {
        return Report{}, fmt.Errorf("unsupported version %q", v)
    }
    r.Version = n

    id, ok := fields["id"]
    if !ok || id == "" {
        return Report{}, fmt.Errorf("missing id")
    }
    r.ID = id

    if s, ok := fields["seq"]; ok {
        r.Seq, err = strconv.ParseUint(s, 10, 64)
        if err != nil {
            return Report{}, fmt.Errorf("bad seq %q: %w", s, err)
        }
    }

    if s, ok := fields["conns"]; ok {
        r.Conns, err = strconv.Atoi(s)
        if err != nil {
            return Report{}, fmt.Errorf("bad conns %q: %w", s, err)
        }
        if r.Conns < 0 {
            return Report{}, fmt.Errorf("negative conns %q", s)
        }
    }

    if s, ok := fields["rps"]; ok {
        r.RPS, err = strconv.ParseFloat(s, 64)
        if err != nil {
            return Report{}, fmt.Errorf("bad rps %q: %w", s, err)
        }
        if math.IsNaN(r.RPS) || math.IsInf(r.RPS, 0) {
            return Report{}, fmt.Errorf("non-finite rps %q", s)
        }
    }

    return r, nil
}
